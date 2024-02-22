package connector

import (
	"context"

	"github.com/LearningMotors/go-genproto/learningmotors/pb/composer"
)

type NotePull interface {
	PullSectionsInfo(ctx context.Context, metadata *composer.Metadata, compositionID string, vcSectionIDs []string) ([]*composer.SectionS2, error)
	PullDiagnosesInfo(ctx context.Context, metadata *composer.Metadata, compositionID string) ([]*composer.SectionS2, error)
}
