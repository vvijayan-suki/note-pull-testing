package main

import (
	"context"
	"fmt"

	"github.com/LearningMotors/go-genproto/learningmotors/appointments"
	"github.com/LearningMotors/go-genproto/learningmotors/pb/composer"
	"github.com/LearningMotors/go-genproto/learningmotors/pb/emrtypes"
	"github.com/LearningMotors/go-genproto/learningmotors/pb/patients"
	"github.com/LearningMotors/go-genproto/learningmotors/pb/user"
	"github.com/LearningMotors/go-genproto/suki/pb/emr"
	"github.com/LearningMotors/platform/service/metadata"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_metadata "google.golang.org/grpc/metadata"

	"github.com/vvijayan-suki/note-pull-testing/source/internal/connector/patientcharts"
)

func main() {

	ctx := context.Background()

	tags := grpc_ctxtags.NewTags()
	tags = tags.Set(metadata.SukiOrganizationID, "a25b45a3-984a-473f-a18d-f1d4abc2e179")

	ctx = grpc_ctxtags.SetInContext(ctx, tags)

	md := grpc_metadata.New(map[string]string{
		metadata.SukiOrganizationID: "a25b45a3-984a-473f-a18d-f1d4abc2e179",
	})

	grpc_metadata.NewOutgoingContext(ctx, md)

	patientChartsConnector := patientcharts.Init(ctx)

	metadata := &composer.Metadata{
		Status:     0,
		NotetypeId: "",
		Name:       "",
		User: &user.User{
			Id:             "8fba7c82-b5c9-4819-be46-3e4e9b274856",
			Roles:          nil,
			Person:         nil,
			AuthId:         "",
			Email:          "",
			Specialty:      "",
			Timezone:       "",
			CreatedAt:      nil,
			UpdatedAt:      nil,
			UserType:       "",
			ActivationSent: nil,
			OrganizationId: "a25b45a3-984a-473f-a18d-f1d4abc2e179",
			Specialties:    nil,
			RegisteredAt:   nil,
			PhoneNumber:    "",
			SharedUser:     false,
			EnhancedReview: 0,
		},
		Patient: &patients.Patient{
			Id:         "00ca5092-0f5f-4c05-8ae0-84ea2adaa47e",
			Person:     nil,
			Mrn:        "6464",
			EmrId:      "6464",
			EmrIdType:  "",
			FhirId:     "",
			FhirIdType: "",
			CreatedAt:  nil,
			UpdatedAt:  nil,
			DeletedAt:  nil,
		},
		Appointment: &appointments.Appointment{
			Id:                    "63deaffd-13ce-4e35-a5d3-4a74b2598939",
			FhirId:                "",
			OwnerId:               "8fba7c82-b5c9-4819-be46-3e4e9b274856",
			PatientId:             "",
			StartsAt:              nil,
			EndsAt:                nil,
			AppointmentType:       0,
			MedicalProcedure:      "",
			Reason:                "",
			EmrId:                 "891252",
			EmrEncounterId:        "46274",
			EmrEncounterStatus:    0,
			FhirEncounterId:       "",
			FhirEncounterIdType:   "",
			EmrDepartmentId:       "150",
			AppointmentLocationId: "",
			EmrUserId:             "",
			NoteIdentifier:        nil,
			CreatedAt:             nil,
			UpdatedAt:             nil,
			DeletedAt:             nil,
			CompositionIds:        nil,
			IsImported:            false,
			StartTimeSetByUser:    false,
			EncounterDetails:      nil,
			NoteIds:               nil,
		},
		EncounterAddress: nil,
		SubmissionInformation: &emr.SubmissionInformation{
			Destinations:       nil,
			TryAllDestinations: false,
			EmrType:            emrtypes.EMRType_ATHENA,
			EmrSecondaryType:   emrtypes.EMRSecondaryType_ATHENA_EMR,
		},
		SubmissionStatus:  nil,
		EmrStatus:         nil,
		ForceSubmit:       false,
		ReviewMessage:     "",
		DateSignedOff:     nil,
		ClientType:        0,
		CreatedWithS2:     false,
		EmrNoteType:       nil,
		PulledNoteFromEmr: false,
		PartialDictation:  nil,
		Source:            nil,
		DateOfService:     nil,
	}

	for _, sectionID := range []string{"a8388523-69a3-495e-a7ab-70057669c5b8"} {
		response, err := patientChartsConnector.PullSectionsInfo(ctx, metadata, "fe79d0c6-ee46-4499-86d2-3484d7c528cc", []string{sectionID})
		fmt.Println(response)
		fmt.Println(err)
	}
}
