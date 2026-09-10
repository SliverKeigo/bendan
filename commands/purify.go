package commands

import (
	"context"
	"net/url"
	"strings"
	"sync"

	"github.com/sxyazi/bendan/commands/purify"
	"github.com/sxyazi/bendan/platform"
	"github.com/sxyazi/bendan/utils"
)

type purifyResult struct {
	before *url.URL
	after  *url.URL
}

func purifyDo(text string) <-chan []*purifyResult {
	urls := utils.ExtractUrls(text)
	todo := make([]*purifyResult, 0, len(urls))
	for _, value := range urls {
		if purify.Tracks.Test(value) {
			todo = append(todo, &purifyResult{before: value})
		}
	}
	if len(todo) == 0 {
		return nil
	}

	result := make(chan []*purifyResult, 1)
	go func() {
		var waitGroup sync.WaitGroup
		waitGroup.Add(len(todo))
		for _, item := range todo {
			go func(item *purifyResult) {
				defer waitGroup.Done()
				clone := *item.before
				item.after = purify.Tracks.Do(&purify.Stage{URL: &clone})
			}(item)
		}
		waitGroup.Wait()
		result <- todo
	}()
	return result
}

// Purify replies with tracking-free URLs. OneBot has no inline-query equivalent.
func Purify(ctx context.Context, message *platform.Message) bool {
	result := purifyDo(message.Content())
	if result == nil {
		return false
	}

	var text strings.Builder
	for _, item := range <-result {
		if item.after != nil {
			text.WriteString(item.after.String())
			text.WriteByte('\n')
		}
	}
	if text.Len() == 0 {
		return false
	}
	urls := strings.TrimSpace(text.String())
	if strings.Count(urls, "\n") == 0 {
		replyText(ctx, message, "净化后的链接：\n"+urls)
	} else {
		replyText(ctx, message, "净化后的链接：\n"+urls)
	}
	return true
}
