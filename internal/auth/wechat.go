package auth

import (
	"context"
	"honi/internal/agent"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"github.com/SpellingDragon/wechat-robot-go/wechat"
	"github.com/mdp/qrterminal/v4"
)

var GetWechatBot = sync.OnceValue(func() *Wechat {
	home, err := os.UserHomeDir()
	if err != nil {
		slog.Error("UserHomeDir failed", "err", err)
		os.Exit(1)
	}
	appHome := filepath.Join(home, ".config", "honi")
	return &Wechat{
		bot: wechat.NewBot(
			wechat.WithTokenFile(filepath.Join(appHome, "weixin-token.json")),
			wechat.WithContextTokenDir(filepath.Join(appHome, "wechat-context-tokens"))),
		ctx: context.Background(),
	}
})

type Wechat struct {
	bot *wechat.Bot
	ctx context.Context
}

func (w *Wechat) HasValidateCredentials() bool {
	w.bot.ClearAllContextTokens()
	return w.bot.HasValidateCredentials()
}

func (w *Wechat) Run() error {
	return w.bot.Run(w.ctx)
}

func (w *Wechat) Login() error {
	ctx := context.Background()
	err := w.bot.Login(ctx, func(qrCodeUrl string) {
		config := qrterminal.Config{
			Writer:     os.Stdout,
			HalfBlocks: true,
			Level:      qrterminal.L,
			QuietZone:  2,
		}
		qrterminal.GenerateWithConfig(qrCodeUrl, config)
	})
	if err != nil {
		slog.Error("微信登录失败", "err", err)
		return err
	}

	// 注册消息处理器
	w.bot.OnMessage(func(ctx context.Context, msg *wechat.Message) error {
		text := msg.Text()
		if text == "" {
			return nil
		}
		answer, err := agent.GetAgent().Ask(text)
		if err != nil {
			return err
		}
		return w.bot.Reply(ctx, msg, answer)
	})

	return nil
}
