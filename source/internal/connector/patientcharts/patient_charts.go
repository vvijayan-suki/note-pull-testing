package patientcharts

import (
	"context"
	"fmt"
	"sync"

	"github.com/LearningMotors/go-genproto/learningmotors/pb/composer"
	"github.com/LearningMotors/go-genproto/learningmotors/pb/emrtypes"
	"github.com/LearningMotors/go-genproto/learningmotors/pb/service"
	patientchartsPB "github.com/LearningMotors/go-genproto/suki/pb/patientcharts"
	"github.com/LearningMotors/platform/clients"
	"github.com/LearningMotors/platform/service/ctxlog"
	"github.com/sourcegraph/conc/pool"
)

var (
	once          sync.Once
	serviceClient *ServiceClient
)

type ServiceClient struct {
	client patientchartsPB.MsPatientChartsClient
}

func Init(ctx context.Context) *ServiceClient {
	once.Do(func() {
		conn := clients.NewClientConn(ctx, service.ServiceName_ms_patient_charts)
		serviceClient = &ServiceClient{
			client: patientchartsPB.NewMsPatientChartsClient(conn),
		}
	})

	return serviceClient
}

type RequestData struct {
	OrganizationID    string
	EmrEncounterID    string
	EmrPatientID      string
	EmrDepartmentID   string
	EmrUserID         string
	SukiAppointmentID string
	EmrSecondaryType  emrtypes.EMRSecondaryType
	FhirPatientID     string
	FhirEncounterID   string
	CompositionID     string
	PatientID         string
}

func (sc *ServiceClient) PullSectionsInfo(ctx context.Context, metadata *composer.Metadata, compositionID string, vcSectionIDs []string) ([]*composer.SectionS2, error) {
	sugar := ctxlog.Sugar(ctx)

	data := sc.calculateRequestData(metadata, compositionID)

	sections := make([]*composer.SectionS2, 0, len(vcSectionIDs))

	if emrtypes.EMRSecondaryType(data.EmrSecondaryType) == emrtypes.EMRSecondaryType_CERNER_EMR {
		request := &patientchartsPB.SyncEMRNoteRequest{
			OrganizationId:    data.OrganizationID,
			EmrEncounterId:    data.EmrEncounterID,
			SukiSectionId:     vcSectionIDs[0],
			EmrPatientId:      data.EmrPatientID,
			EmrDepartmentId:   data.EmrDepartmentID,
			EmrUserId:         data.EmrUserID,
			SukiSectionIds:    vcSectionIDs,
			SukiAppointmentId: data.SukiAppointmentID,
			CompositionId:     data.CompositionID,
		}
		response, err := sc.client.SyncEMRNote(ctx, request)
		if err != nil {
			return nil, fmt.Errorf("error in SyncEMRNote: %v", err)
		}

		if response.GetStatus() != patientchartsPB.SyncEMRNoteResponse_UPDATE {
			err = sc.handleStatus(ctx, request, response)
			if err != nil {
				return nil, err
			}
			return nil, nil
		}
		sugar.Infof("successfully updated sections from SyncEMRNote")

		returnedSectionIDs := response.GetSukiSectionIds()
		for i, sectionContent := range response.GetSectionsContent() {
			sections = append(sections, &composer.SectionS2{
				Id:        returnedSectionIDs[i],
				ContentS2: sectionContent,
			})
		}

		return sections, nil
	}

	/* conc.pool from the conc library uses go routines inside for parallelization, but
	has the added advantage of simplifying our calls by collecting the intended result
	from each loop (SectionS2) and stops all go routines on the first error, which it then returns
	*/
	p := pool.NewWithResults[*composer.SectionS2]().WithContext(ctx).WithCancelOnError().WithFirstError()

	for _, sectionID := range vcSectionIDs {
		sectionID := sectionID
		p.Go(func(ctx context.Context) (*composer.SectionS2, error) {
			return sc.pullSection(ctx, data, sectionID)
		})
	}

	sections, err := p.Wait()
	if err != nil {
		return nil, err
	}

	return sections, nil
}

func (sc *ServiceClient) pullSection(ctx context.Context, data RequestData, sectionID string) (*composer.SectionS2, error) {
	sugar := ctxlog.Sugar(ctx)

	request := &patientchartsPB.SyncEMRNoteRequest{
		OrganizationId:    data.OrganizationID,
		EmrEncounterId:    data.EmrEncounterID,
		SukiSectionId:     sectionID,
		EmrPatientId:      data.EmrPatientID,
		EmrDepartmentId:   data.EmrDepartmentID,
		EmrUserId:         data.EmrUserID,
		SukiAppointmentId: data.SukiAppointmentID,
		CompositionId:     data.CompositionID,
	}
	response, err := sc.client.SyncEMRNote(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("for composition id: %s, section id: %s, error in SyncEMRNote: %v", data.CompositionID, sectionID, err)
	}

	if response.GetStatus() != patientchartsPB.SyncEMRNoteResponse_UPDATE {
		err = sc.handleStatus(ctx, request, response)
		if err != nil {
			return nil, err
		}

		return nil, nil
	}
	sugar.Infof("successfully updated sections from SyncEMRNote")
	return &composer.SectionS2{
		Id:        sectionID,
		ContentS2: response.GetSectionContent(),
	}, nil
}

func (sc *ServiceClient) PullDiagnosesInfo(ctx context.Context, metadata *composer.Metadata, compositionID string) ([]*composer.SectionS2, error) {
	sugar := ctxlog.Sugar(ctx)

	data := sc.calculateRequestData(metadata, compositionID)

	request := &patientchartsPB.GetEncounterDiagnosesSectionsRequest{
		OrganizationId:    data.OrganizationID,
		SukiPatientId:     data.PatientID,
		EmrEncounterId:    data.EmrEncounterID,
		EmrPatientId:      data.EmrPatientID,
		FhirPatientId:     data.FhirPatientID,
		FhirEncounterId:   data.FhirEncounterID,
		SukiAppointmentId: data.SukiAppointmentID,
	}

	response, err := sc.client.GetEncounterDiagnosesSections(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("error in GetEncounterDiagnosesSections : %v", err)
	}

	if response.GetStatus() != patientchartsPB.GetEncounterDiagnosesSectionsResponse_SUCCESS {
		return nil, fmt.Errorf("error in GetEncounterDiagnosesSections : %v", response.GetStatus())
	}

	sugar.Infof("successfully received encounters from GetEncounterDiagnosesSections")

	return response.GetSections(), nil
}

func (sc *ServiceClient) handleStatus(ctx context.Context, request *patientchartsPB.SyncEMRNoteRequest, response *patientchartsPB.SyncEMRNoteResponse) error {
	sugar := ctxlog.Sugar(ctx)
	switch response.GetStatus() {
	case patientchartsPB.SyncEMRNoteResponse_ERROR:
		return fmt.Errorf("error in SyncEMRNote: %v", response.GetError())
	case patientchartsPB.SyncEMRNoteResponse_NO_UPDATE:
		sugar.Warnf("empty section(s) in EMR Note for EMR Encounter ID: %s", request.GetEmrEncounterId())
	case patientchartsPB.SyncEMRNoteResponse_NO_ENCOUNTER:
		sugar.Warnf("no encounter found in EMR to retrieve Section Content from, for composition: %s", request.GetCompositionId())
	case patientchartsPB.SyncEMRNoteResponse_NO_EMR_NOTE:
		sugar.Warnf("no EMR Note found in EMR to retrieve Section Content from, for composition: %s", request.GetCompositionId())
	case patientchartsPB.SyncEMRNoteResponse_UNSUPPORTED:
		sugar.Warnf("section-based note pull is unsupported for EMR Organization: %s", request.GetOrganizationId())
	default:
		sugar.Warnf("unknown status in SyncEMRNote: %v", response.GetStatus())
	}
	return nil
}

func (sc *ServiceClient) calculateRequestData(metadata *composer.Metadata, compositionID string) RequestData {
	organizationID := metadata.GetUser().GetOrganizationId()
	emrEncounterID := metadata.GetAppointment().GetEmrEncounterId()
	emrPatientID := metadata.GetPatient().GetEmrId()
	emrDepartmentID := metadata.GetAppointment().GetEmrDepartmentId()
	emrUserID := metadata.GetAppointment().GetOwnerId()
	sukiAppointmentID := metadata.GetAppointment().GetId()
	emrSecondaryType := metadata.GetSubmissionInformation().GetEmrSecondaryType()
	fhirPatientID := metadata.GetPatient().GetFhirId()
	fhirEncounterID := metadata.GetAppointment().GetFhirEncounterId()
	patientID := metadata.GetPatient().GetId()

	return RequestData{
		OrganizationID:    organizationID,
		EmrEncounterID:    emrEncounterID,
		EmrPatientID:      emrPatientID,
		EmrDepartmentID:   emrDepartmentID,
		EmrUserID:         emrUserID,
		SukiAppointmentID: sukiAppointmentID,
		EmrSecondaryType:  emrSecondaryType,
		FhirPatientID:     fhirPatientID,
		FhirEncounterID:   fhirEncounterID,
		CompositionID:     compositionID,
		PatientID:         patientID,
	}
}
