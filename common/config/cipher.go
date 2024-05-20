package config

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/hashicorp/go-hclog"
)

var (
	region    = "ap-southeast-1"
	awsConfig = &aws.Config{
		Region: &region,
	}
)

type Cipher struct {
	Decrypt func(string, []byte) (string, error)
}

func awsKms() *kms.KMS {
	sess := session.Must(session.NewSession(awsConfig))
	creds := credentials.NewChainCredentials(
		[]credentials.Provider{
			&credentials.EnvProvider{},
			&ec2rolecreds.EC2RoleProvider{
				Client: ec2metadata.New(sess),
			},
		})
	return kms.New(sess, &aws.Config{Credentials: creds})
}

func KmsDecrypt(keyId string, ciphertext []byte) (string, error) {
	input := &kms.DecryptInput{
		KeyId:          &keyId,
		CiphertextBlob: ciphertext,
	}
	output, err := awsKms().Decrypt(input)
	if err != nil {
		hclog.L().Error("decrypting ciphertext via aws kms failed", "err", err)
		return "", err
	}
	return string(output.Plaintext), nil
}

func KmsEncrypt(keyId string, plaintext []byte) ([]byte, error) {
	input := &kms.EncryptInput{
		KeyId:     &keyId,
		Plaintext: plaintext,
	}
	output, err := awsKms().Encrypt(input)
	if err != nil {
		hclog.L().Error("encrypting plaintext via aws kms failed", "err", err)
		return nil, err
	}
	return output.CiphertextBlob, nil
}
