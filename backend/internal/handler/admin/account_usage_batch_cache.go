package admin

import (
	"strconv"
	"strings"
	"time"
)

var accountUsageBatchCache = newSnapshotCache(30 * time.Second)

func buildAccountUsageBatchCacheKey(accountIDs []int64) string {
	if len(accountIDs) == 0 {
		return "accounts_usage_empty"
	}
	var b strings.Builder
	b.Grow(len(accountIDs) * 6)
	_, _ = b.WriteString("accounts_usage:")
	for i, id := range accountIDs {
		if i > 0 {
			_ = b.WriteByte(',')
		}
		_, _ = b.WriteString(strconv.FormatInt(id, 10))
	}
	return b.String()
}
