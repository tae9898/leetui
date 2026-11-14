package stmt

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"leetui/internal/ps/lcref"
)

func questionPath(root, site, slug string) string {
	return filepath.Join(root, ".stmt", site, slug+".json")
}

// CachedQuestion은 문제 메타를 가져오되 .stmt/lc/<slug>.json 캐시를 먼저 읽는다.
// 캐시 미스면 fetch해 json(전체 Question)과 html(지문 원문)을 함께 남긴다.
func CachedQuestion(ctx context.Context, root, site, slug string) (lcref.Question, error) {
	qp := questionPath(root, site, slug)
	if data, err := os.ReadFile(qp); err == nil {
		var q lcref.Question
		if json.Unmarshal(data, &q) == nil && q.Content != "" {
			return q, nil
		}
	}
	q, err := lcref.FetchQuestion(ctx, slug)
	if err != nil {
		return lcref.Question{}, err
	}
	if err := os.MkdirAll(filepath.Dir(qp), 0o755); err != nil {
		return lcref.Question{}, err
	}
	if data, err := json.Marshal(q); err == nil {
		_ = os.WriteFile(qp, data, 0o644)
	}
	_ = os.WriteFile(htmlPath(root, site, slug), []byte(q.Content), 0o644)
	return q, nil
}
