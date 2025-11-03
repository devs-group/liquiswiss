package api_service_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"liquiswiss/internal/service/api_service"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
)

func TestCalculateSalaryAdjustmentsCountsBothDistribution(t *testing.T) {
	service := &api_service.APIService{}

	costs := []models.SalaryCost{
		{
			DistributionType: "employee",
			AmountType:       "fixed",
			Cycle:            utils.CycleMonthly,
			CalculatedAmount: 10000,
		},
		{
			DistributionType: "employer",
			AmountType:       "fixed",
			Cycle:            utils.CycleMonthly,
			CalculatedAmount: 20000,
		},
		{
			DistributionType: "both",
			AmountType:       "fixed",
			Cycle:            utils.CycleMonthly,
			CalculatedAmount: 30000,
		},
		{
			DistributionType: "both",
			AmountType:       "percentage",
			Cycle:            utils.CycleMonthly,
			CalculatedAmount: 15000,
		},
	}

	employeeTotal := service.CalculateSalaryAdjustments(utils.CycleMonthly, "employee", costs)
	require.Equal(t, uint64(55000), employeeTotal)

	employerTotal := service.CalculateSalaryAdjustments(utils.CycleMonthly, "employer", costs)
	require.Equal(t, uint64(65000), employerTotal)
}
