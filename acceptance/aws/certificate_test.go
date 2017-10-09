package awsat

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/acm"
)

func TestCertificate(t *testing.T) {
	t.Run("create", func(t *testing.T) {
		Template("create certificate domains=test1.com,subdomain.test2.com validation-domains=test1.com,test2.com").
			Mock(&acmMock{
				RequestCertificateFunc: func(param0 *acm.RequestCertificateInput) (*acm.RequestCertificateOutput, error) {
					return &acm.RequestCertificateOutput{CertificateArn: String("arn:my:new:certificate")}, nil
				},
			}).ExpectInput("RequestCertificate", &acm.RequestCertificateInput{
			DomainName:              String("test1.com"),
			SubjectAlternativeNames: []*string{String("subdomain.test2.com")},
			DomainValidationOptions: []*acm.DomainValidationOption{
				{DomainName: String("test1.com"), ValidationDomain: String("test1.com")},
				{DomainName: String("subdomain.test2.com"), ValidationDomain: String("test2.com")},
			},
		}).ExpectCommandResult("arn:my:new:certificate").ExpectCalls("RequestCertificate").Run(t)
	})

	t.Run("delete", func(t *testing.T) {
		Template("delete certificate arn=arn:certificate:to:delete").
			Mock(&acmMock{
				DeleteCertificateFunc: func(param0 *acm.DeleteCertificateInput) (*acm.DeleteCertificateOutput, error) {
					return nil, nil
				},
			}).ExpectInput("DeleteCertificate", &acm.DeleteCertificateInput{CertificateArn: String("arn:certificate:to:delete")}).
			ExpectCalls("DeleteCertificate").Run(t)
	})

	t.Run("check", func(t *testing.T) {
		Template("check certificate arn=arn:certificate:to:check state=issued timeout=1").Mock(&acmMock{
			DescribeCertificateFunc: func(input *acm.DescribeCertificateInput) (*acm.DescribeCertificateOutput, error) {
				return &acm.DescribeCertificateOutput{Certificate: &acm.CertificateDetail{CertificateArn: String("arn:certificate:to:check"), Status: String("issued")}}, nil
			}}).ExpectInput("DescribeCertificate", &acm.DescribeCertificateInput{CertificateArn: String("arn:certificate:to:check")}).
			ExpectCalls("DescribeCertificate").Run(t)
	})
}
