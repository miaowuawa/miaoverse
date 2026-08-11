package Article

import (
	"context"
	"errors"
	"sync"

	"miaoverse/consts"
	"miaoverse/dao/article"
	modelarticle "miaoverse/model/dao/article"
	modeluser "miaoverse/model/dao/user"
	"miaoverse/model/dto/resp"
	"miaoverse/model/server"
)

// DetailResult 文章详情组装结果：正文（含截断/分段处理）+ 互动计数。
// Content 为最终交付给调用方的正文（可能被 60% 截断或截取首段）；
// Full 表示是否交付完整正文；Segments 表示交付正文被切分的分段数（0 表示无需分段）。
type DetailResult struct {
	Stats    *modelarticle.ArticleInteractCount
	Content  string
	Full     bool
	Segments int
}

// BuildDetail 跨库组装文章详情（正文取自 MongoDB，互动计数取自 MySQL，两库并行拉取）。
// meta 由中间件解析并缓存，此处不再回查 MySQL 元数据，避免重复跨库。
// 匿名可见性规则（与 BuildSegment 共用同一口径）：
//   - 普通文章/小说根文章（chapter_id=0）：未登录仅可见正文前 60%；
//   - 小说章节：未登录仅支持前 AnonymousNovelMaxChapter 章，其余章节不返回正文（不触达 MongoDB）。
func BuildDetail(ctx context.Context, servants *server.Servants, meta *modelarticle.Metadata, loggedIn bool) (*DetailResult, error) {
	needBody := loggedIn || meta.ChapterID <= consts.AnonymousNovelMaxChapter

	body, stats, err := fetchBodyAndStats(ctx, servants, meta, needBody)
	if err != nil {
		return nil, err
	}

	content, full := deliverableContent(body.Content, meta, loggedIn)
	result := &DetailResult{
		Stats:   stats,
		Content: content,
		Full:    full,
	}
	if segments := segmentCount(content); segments > 1 {
		result.Segments = segments
		result.Content = firstSegment(content)
	}
	return result, nil
}

// BuildSegment 跨库拉取文章正文的指定分段（seq 从 1 起）。
// 与 BuildDetail 共用截断/分段口径：未登录普通文章前 60%、小说前 2 章可交付。
// 返回 ok=false 表示 seq 超出交付正文的分段范围（handler 视为参数错误）。
func BuildSegment(ctx context.Context, servants *server.Servants, meta *modelarticle.Metadata, seq int, loggedIn bool) (resp.ArticleSegment, bool, error) {
	if seq < 1 || seq > consts.ArticleSegmentsMaxSeq {
		return resp.ArticleSegment{}, false, nil
	}

	needBody := loggedIn || meta.ChapterID <= consts.AnonymousNovelMaxChapter
	body, _, err := fetchBodyAndStats(ctx, servants, meta, needBody)
	if err != nil {
		return resp.ArticleSegment{}, false, err
	}

	content, full := deliverableContent(body.Content, meta, loggedIn)
	if seq > segmentCount(content) {
		return resp.ArticleSegment{}, false, nil
	}
	return resp.ArticleSegment{
		ID:      meta.ID,
		Seq:     seq,
		Content: segmentSlice(content, seq),
		Full:    full,
	}, true, nil
}

// ToArticleInfo 将文章元数据 + 组装结果 + 作者/互动状态转换为详情响应 DTO。
func ToArticleInfo(meta *modelarticle.Metadata, result *DetailResult, author *modeluser.User, isLiked bool, isFollowing bool) resp.ArticleInfo {
	info := resp.ArticleInfo{
		ID:          meta.ID,
		UserID:      meta.UserID,
		Title:       meta.Title,
		Description: meta.Description,
		Cover:       meta.Cover,
		Type:        meta.Type,
		NovelID:     meta.NovelID,
		ChapterID:   meta.ChapterID,
		Content:     result.Content,
		Status:      meta.Status,
		CreatedAt:   meta.CreatedAt,
		UpdatedAt:   meta.UpdatedAt,
		IsLiked:     isLiked,
		IsFollowing: isFollowing,
		Full:        result.Full,
		Segments:    result.Segments,
	}
	if author != nil {
		info.Author = *author
	}
	if result.Stats != nil {
		info.Stats = resp.ArticleStats{
			Likes:    result.Stats.LikeCount,
			Comments: result.Stats.CommentCount,
			Shares:   result.Stats.ShareCount,
			Views:    result.Stats.ViewCount,
		}
	}
	return info
}

// fetchBodyAndStats 并行拉取 MongoDB 正文与 MySQL 互动计数（跨库查询避免串行等待）。
// needBody 为 false 时（未登录访问不可见章节）跳过 MongoDB 查询，正文置空。
// 互动计数行缺失不视为错误（与动态详情口径一致），stats 返回 nil。
func fetchBodyAndStats(ctx context.Context, servants *server.Servants, meta *modelarticle.Metadata, needBody bool) (*modelarticle.Article, *modelarticle.ArticleInteractCount, error) {
	var (
		wg    sync.WaitGroup
		mu    sync.Mutex
		first error
		body  *modelarticle.Article
		stats *modelarticle.ArticleInteractCount
	)
	recordErr := func(err error) {
		if err == nil {
			return
		}
		mu.Lock()
		if first == nil {
			first = err
		}
		mu.Unlock()
	}

	if needBody {
		wg.Add(1)
		go func() {
			defer wg.Done()
			b, err := servants.ArticleServant.QueryBodyByMeta(ctx, meta)
			if err != nil {
				recordErr(err)
				return
			}
			body = b
		}()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		s, err := servants.ArticleServant.QueryInteractCount(meta.ID)
		if err != nil {
			if !errors.Is(err, article.ErrArticleNotFound) {
				recordErr(err)
			}
			return
		}
		stats = s
	}()

	wg.Wait()
	if first != nil {
		return nil, nil, first
	}
	if body == nil {
		body = &modelarticle.Article{}
	}
	return body, stats, nil
}

// deliverableContent 计算当前身份可交付的正文与是否完整。
// 未登录：非章节（普通文章/小说根）截断前 60%；小说章节超过前 2 章返回空正文（Full=false）。
// 注意：正文按 rune 处理，避免按字节截断破坏 UTF-8 字符。
func deliverableContent(content string, meta *modelarticle.Metadata, loggedIn bool) (string, bool) {
	if loggedIn {
		return content, true
	}
	if meta.ChapterID > consts.AnonymousNovelMaxChapter {
		return "", false
	}
	if meta.ChapterID == 0 {
		return truncatePercent(content, consts.AnonymousContentRatio), false
	}
	return content, true
}

// truncatePercent 按百分比截断字符串（rune 安全）。
func truncatePercent(s string, percent int) string {
	if percent <= 0 {
		return ""
	}
	runes := []rune(s)
	n := len(runes) * percent / 100
	if n >= len(runes) {
		return s
	}
	return string(runes[:n])
}

// segmentSize 单个正文分段的 rune 上限（ArticleSegmentSize 千字）。
func segmentSize() int {
	return consts.ArticleSegmentSize * 1000
}

// segmentCount 计算正文所需分段数；正文长度不超过一个分段时返回 0（无需分段）。
func segmentCount(s string) int {
	size := segmentSize()
	n := len([]rune(s))
	if n <= size {
		return 0
	}
	return (n + size - 1) / size
}

// firstSegment 返回正文首段（长度上限为 segmentSize）。
func firstSegment(s string) string {
	return segmentSlice(s, 1)
}

// segmentSlice 返回正文第 seq 段（seq 从 1 起，rune 安全切片）。
func segmentSlice(s string, seq int) string {
	runes := []rune(s)
	size := segmentSize()
	start := (seq - 1) * size
	if start >= len(runes) {
		return ""
	}
	end := seq * size
	if end > len(runes) {
		end = len(runes)
	}
	return string(runes[start:end])
}
