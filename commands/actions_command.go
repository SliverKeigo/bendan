package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/sxyazi/bendan/platform"
)

func Actions(ctx context.Context, message *platform.Message) bool {
	params := strings.Fields(message.Text)
	if len(params) == 0 || params[0] != "//actions" {
		return false
	}
	if !isAdministrator(message) {
		return true
	}

	if len(params) == 1 {
		path, updated, zh, latin := actionLexiconStatus()
		replyText(ctx, message, fmt.Sprintf("动作词表\n路径：%s\n中文动作：%d\n英文别名：%d\n加载时间：%s", path, zh, latin, updated.Format("2006-01-02 15:04:05")))
		return true
	}
	if len(params) == 2 && params[1] == "list" {
		lexicon := actionLexiconSnapshot()
		replyText(ctx, message, fmt.Sprintf("中文：%s\n英文：%s", strings.Join(sortedKeys(lexicon.Chinese), "、"), strings.Join(sortedKeys(lexicon.Latin), "、")))
		return true
	}
	if len(params) == 2 && params[1] == "reload" {
		if err := LoadActionLexicon(actionLexiconPath()); err != nil {
			replyText(ctx, message, "词表重载失败："+err.Error())
			return true
		}
		zh, latin := actionLexiconCounts()
		replyText(ctx, message, fmt.Sprintf("词表已重载：中文 %d，英文 %d", zh, latin))
		return true
	}
	if len(params) >= 5 && params[1] == "add" {
		kind, action, display := params[2], params[3], strings.Join(params[4:], " ")
		if err := updateActionLexicon(kind, action, display); err != nil {
			replyText(ctx, message, "添加动作失败："+err.Error())
			return true
		}
		replyText(ctx, message, fmt.Sprintf("已添加 %s 动作：%s -> %s", kind, action, display))
		return true
	}
	if len(params) == 4 && params[1] == "remove" {
		kind, action := params[2], params[3]
		if err := removeActionLexiconEntry(kind, action); err != nil {
			replyText(ctx, message, "删除动作失败："+err.Error())
			return true
		}
		replyText(ctx, message, fmt.Sprintf("已删除 %s 动作：%s", kind, action))
		return true
	}
	replyText(ctx, message, "用法：//actions [list|reload]；//actions add <zh|latin> <动作> <输出>；//actions remove <zh|latin> <动作>")
	return true
}

func updateActionLexicon(kind, action, display string) error {
	lexicon := clonedActionLexicon(actionLexiconSnapshot())
	values, err := actionLexiconEntries(&lexicon, kind)
	if err != nil {
		return err
	}
	values[action] = display
	return saveActionLexicon(lexicon)
}

func removeActionLexiconEntry(kind, action string) error {
	lexicon := clonedActionLexicon(actionLexiconSnapshot())
	values, err := actionLexiconEntries(&lexicon, kind)
	if err != nil {
		return err
	}
	if _, ok := values[action]; !ok {
		return fmt.Errorf("动作 %q 不存在", action)
	}
	delete(values, action)
	if err := lexicon.validate(); err != nil {
		return err
	}
	return saveActionLexicon(lexicon)
}

func clonedActionLexicon(source actionLexicon) actionLexicon {
	return actionLexicon{
		Chinese: cloneActionEntries(source.Chinese),
		Latin:   cloneActionEntries(source.Latin),
	}
}

func cloneActionEntries(source map[string]string) map[string]string {
	result := make(map[string]string, len(source))
	for action, display := range source {
		result[action] = display
	}
	return result
}

func actionLexiconEntries(lexicon *actionLexicon, kind string) (map[string]string, error) {
	switch kind {
	case "zh":
		if lexicon.Chinese == nil {
			lexicon.Chinese = make(map[string]string)
		}
		return lexicon.Chinese, nil
	case "latin":
		if lexicon.Latin == nil {
			lexicon.Latin = make(map[string]string)
		}
		return lexicon.Latin, nil
	default:
		return nil, fmt.Errorf("类型必须为 zh 或 latin")
	}
}

func saveActionLexicon(lexicon actionLexicon) error {
	if err := lexicon.validate(); err != nil {
		return err
	}
	path, _, _, _ := actionLexiconStatus()
	data, err := json.MarshalIndent(lexicon, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化词表: %w", err)
	}
	data = append(data, '\n')

	directory := filepath.Dir(path)
	temporary, err := os.CreateTemp(directory, ".actions-*.json")
	if err != nil {
		return fmt.Errorf("创建临时词表: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if _, err := temporary.Write(data); err != nil {
		temporary.Close()
		return fmt.Errorf("写入临时词表: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("关闭临时词表: %w", err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("替换词表: %w", err)
	}
	if err := LoadActionLexicon(path); err != nil {
		return fmt.Errorf("重载已保存词表: %w", err)
	}
	return nil
}

func sortedKeys(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
