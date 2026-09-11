package commands

import (
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/sxyazi/bendan/utils"
)

//go:embed actions.default.json
var defaultActionData []byte

const defaultActionLexiconPath = "actions.json"

type actionLexicon struct {
	Chinese map[string]string `json:"zh"`
	Latin   map[string]string `json:"latin"`
}

var actionState = struct {
	sync.RWMutex
	lexicon actionLexicon
	path    string
	updated time.Time
}{
	path: defaultActionLexiconPath,
}

func init() {
	if err := LoadActionLexicon(actionLexiconPath()); err != nil {
		panic(fmt.Sprintf("load action lexicon: %v", err))
	}
}

func actionLexiconPath() string {
	if path := utils.Config("action_lexicon_path"); path != "" {
		return path
	}
	return defaultActionLexiconPath
}

// LoadActionLexicon loads a JSON lexicon from path, falling back to the embedded defaults when absent.
func LoadActionLexicon(path string) error {
	lexicon, err := readActionLexicon(path)
	if err != nil {
		return err
	}

	actionState.Lock()
	actionState.lexicon = lexicon
	actionState.path = path
	actionState.updated = time.Now()
	actionState.Unlock()
	return nil
}

func readActionLexicon(path string) (actionLexicon, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return actionLexicon{}, fmt.Errorf("read %s: %w", path, err)
		}
		data = defaultActionData
	}

	var lexicon actionLexicon
	if err := json.Unmarshal(data, &lexicon); err != nil {
		return actionLexicon{}, fmt.Errorf("parse %s: %w", path, err)
	}
	if err := lexicon.validate(); err != nil {
		return actionLexicon{}, err
	}
	return lexicon, nil
}

func WatchActionLexicon(ctx context.Context, interval time.Duration) {
	path, _, _, _ := actionLexiconStatus()
	lastFingerprint := actionLexiconFingerprint(path)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			fingerprint := actionLexiconFingerprint(path)
			if fingerprint == "" || fingerprint == lastFingerprint {
				continue
			}
			if err := LoadActionLexicon(path); err != nil {
				log.Printf("action lexicon reload failed path=%q error=%v; keeping previous lexicon", path, err)
				lastFingerprint = fingerprint
				continue
			}
			lastFingerprint = fingerprint
			zh, latin := actionLexiconCounts()
			log.Printf("action lexicon reloaded path=%q zh=%d latin=%d", path, zh, latin)
		}
	}
}

func actionLexiconFingerprint(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%x", sha256.Sum256(data))
}

func actionLexiconCounts() (zh, latin int) {
	actionState.RLock()
	defer actionState.RUnlock()
	return len(actionState.lexicon.Chinese), len(actionState.lexicon.Latin)
}

func actionLexiconSnapshot() actionLexicon {
	actionState.RLock()
	defer actionState.RUnlock()
	return actionState.lexicon
}

func ActionLexiconStatus() (path string, updated time.Time, zh, latin int) {
	return actionLexiconStatus()
}

func actionLexiconStatus() (path string, updated time.Time, zh, latin int) {
	actionState.RLock()
	defer actionState.RUnlock()
	return actionState.path, actionState.updated, len(actionState.lexicon.Chinese), len(actionState.lexicon.Latin)
}

func (lexicon actionLexicon) validate() error {
	if len(lexicon.Chinese) == 0 && len(lexicon.Latin) == 0 {
		return fmt.Errorf("action lexicon must define zh or latin actions")
	}
	for action, display := range lexicon.Chinese {
		if strings.TrimSpace(action) == "" || strings.TrimSpace(display) == "" {
			return fmt.Errorf("zh action and display must not be empty")
		}
	}
	for action, display := range lexicon.Latin {
		if strings.TrimSpace(action) == "" || strings.TrimSpace(display) == "" {
			return fmt.Errorf("latin action and display must not be empty")
		}
	}
	return nil
}

func isAction(action string) bool {
	if action == "" {
		return false
	}
	lexicon := actionLexiconSnapshot()
	if _, ok := lexicon.Chinese[action]; ok {
		return true
	}
	if strings.HasSuffix(action, "了") {
		_, ok := lexicon.Chinese[strings.TrimSuffix(action, "了")]
		return ok
	}
	_, ok := lexicon.Latin[strings.ToLower(action)]
	return ok
}

func isSlashAction(action string) bool {
	if isAction(action) {
		return true
	}
	runes := []rune(action)
	return len(runes) > 0 && !unicode.IsLetter(runes[0]) && !unicode.IsNumber(runes[0])
}

func actionDisplay(action string) string {
	lexicon := actionLexiconSnapshot()
	if display, ok := lexicon.Chinese[action]; ok {
		return display
	}
	if strings.HasSuffix(action, "了") {
		if _, ok := lexicon.Chinese[strings.TrimSuffix(action, "了")]; ok {
			return action
		}
	}
	if display, ok := lexicon.Latin[strings.ToLower(action)]; ok {
		return display
	}
	return action
}
