// Package qris implements QRIS generation use case.
package qris

import (
	"context"
	"fmt"
	"time"

	"github.com/exernia/botjanweb/internal/application/service"
	"github.com/exernia/botjanweb/internal/domain"
	"github.com/exernia/botjanweb/internal/domain/entity"
)

var _ usecase.QrisUseCase = (*UseCase)(nil)

// UseCase implements QrisUseCase.
type UseCase struct {
	generator usecase.QrisGeneratorPort
	baseQRIS  string
}

// New creates a new QRIS use case.
func New(generator usecase.QrisGeneratorPort, baseQRIS string) *UseCase {
	return &UseCase{
		generator: generator,
		baseQRIS:  baseQRIS,
	}
}

// GenerateQRIS creates a dynamic QRIS and returns the result.
func (uc *UseCase) GenerateQRIS(ctx context.Context, cmd *entity.QrisCommand, msg *entity.Message) (*entity.QrisResult, error) {
	qrisString, imageData, err := uc.generator.GenerateDynamicQRIS(uc.baseQRIS, cmd.Amount, cmd.Deskripsi)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrQrisGeneration, err)
	}

	return &entity.QrisResult{
		QrisString: qrisString,
		ImageData:  imageData,
		Amount:     cmd.Amount,
		Deskripsi:  cmd.Deskripsi,
	}, nil
}

// BuildPending creates a PendingPayment from message and result.
// msgID is the ID of the QRIS image message sent by the bot.
func BuildPending(msg *entity.Message, result *entity.QrisResult, msgID string) *entity.PendingPayment {
	return &entity.PendingPayment{
		MessageID:         msgID,
		OriginalMessageID: msg.ID,
		ChatID:            msg.ChatID,
		SenderJID:         msg.SenderID,
		SenderPhone:       msg.SenderPhone,
		Amount:            result.Amount,
		Deskripsi:         result.Deskripsi,
		CreatedAt:         time.Now(),
	}
}
