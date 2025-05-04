package dtos

type AdminDashboardDTO struct {
	TotalActiveUrls int

	TotalActiveUsers    int
	TotalUsers          int
	TotalSuspendedUsers int

	TotalActiveCustomDomains    int
	TotalActiveCustomDomainUrls int
}

type AdminUserDashboardDTO struct {
	TotalActiveUrls             int
	TotalActiveCustomDomains    int
	TotalActiveCustomDomainUrls int
}
