package token

import "context"

type Token struct {
	Token    string `json:"token"`
	Platform string `json:"platform"`
}

type Tokens []Token

type Storage interface {
	SaveToken(ctx context.Context, userID uint64, token Token) error
	UserTokens(ctx context.Context, userID uint64) (Tokens, error)
	DeleteTokens(ctx context.Context, tokens []string) error
}

func (ts Tokens) Values() []string {
	ret := make([]string, 0, len(ts))
	for _, t := range ts {
		ret = append(ret, t.Token)
	}

	return ret
}
