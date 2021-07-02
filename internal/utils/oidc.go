package utils

import (
	"context"
	"crypto/sha1"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"github.com/keikoproj/iam-manager/api/v1alpha1"
	"github.com/keikoproj/iam-manager/constants"
	"github.com/keikoproj/iam-manager/pkg/log"

	"net/url"
)

//GetIdpServerCertThumbprint gets the Thumbbprint of the certificate which will be used to generate OIDC tokens
//This was taken from AWS repo https://github.com/aws/containers-roadmap/issues/23#issuecomment-530887531 comment
// https://play.golang.org/p/iSobu11ahUi
func GetIdpServerCertThumbprint(ctx context.Context, url string) (string, error) {
	log := log.Logger(ctx, "internal.utils.oidc", "GetIdpServerCertThumbprint")
	log.Info("Calculating Idp Server cert Thumbprint")

	thumbprint := ""
	hostName, err := parseURL(ctx, url)
	if err != nil {
		log.Error(err, "Unable to get the host")
		return thumbprint, err
	}
	conn, err := tls.Dial("tcp", hostName, &tls.Config{
		InsecureSkipVerify: true,
	})
	if err != nil {
		log.Error(err, "Unable to dial remote host")
		return thumbprint, err
	}
	//Close the connection
	defer conn.Close()

	cs := conn.ConnectionState()
	numCerts := len(cs.PeerCertificates)
	var root *x509.Certificate
	// Important! Get the last cert in the chain, which is the root CA.
	if numCerts >= 1 {
		root = cs.PeerCertificates[numCerts-1]
	} else {
		log.Error(err, "Error getting cert list from connection for Idp Cert Thumbprint calculation")
		return thumbprint, err
	}
	thumbprint = fmt.Sprintf("%x", sha1.Sum(root.Raw))
	// print out the fingerprint
	log.Info("Successfully able to retrieve Idp Server cert thumbprint", "thumbprint", thumbprint)
	return thumbprint, nil
}

//parseURL verifies the url and returns hostname and port
func parseURL(ctx context.Context, idpUrl string) (string, error) {
	log := log.Logger(ctx, "internal.utils.oidc", "parseURL")
	resp, err := url.Parse(idpUrl)
	if err != nil {
		log.Error(err, "unable to parse the idp url")
		return "", err
	}

	if resp.Scheme != "https" {
		log.Error(errors.New("OIDC IDP url must start with https"), "OIDC IDP url must start with https", "obtained", resp.Scheme)
		return "", err
	}

	port := resp.Port()

	if resp.Port() == "" {
		port = "443"
	}
	hostName := fmt.Sprintf("%s:%s", resp.Host, port)
	log.Info("url parsed successfully", "hostName", hostName)
	return hostName, nil
}

// ParseIRSAAnnotation parses IAM role to see if the role to be used in IRSA method.
//
// Returns back both a boolean to indicate whether or not the IRSA system should be used,
// and also returns back the name of the ServiceAccount that should be referenced based
// on the annotation value.
func ParseIRSAAnnotation(ctx context.Context, iamRole *v1alpha1.Iamrole) (bool, string) {
	log := log.Logger(ctx, "internal.utils.oidc", "ParseIRSAAnnotation")
	value, err := getAnnotation(ctx, constants.IRSAAnnotation, iamRole.Annotations)
	if err != nil {
		log.V(1).Info("IRSA Annotation not found", "err", err)
		return false, ""
	}
	return true, value
}
