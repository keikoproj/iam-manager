package v1alpha1

import "testing"

func TestTrustPolicyStatement_Checksum(t *testing.T) {
	type fields struct {
		Sid       string
		Effect    Effect
		Action    string
		Principal Principal
		Condition *Condition
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "test1",
			fields: fields{
				Sid:    "test1",
				Effect: "Allow",
				Action: "sts:AssumeRoleWithWebIdentity",
				Principal: Principal{
					Federated: "arn:aws:iam::123456789012:oidc-provider/oidc.eks.us-east-2.amazonaws.com/id/EXAMPLED539D4633E53DE1B716D3041",
				},
			},
			want: "c0b4c0c0",
		},
		{
			name: "test2: empty sid",
			fields: fields{
				Sid:    "",
				Effect: "Allow",
				Action: "sts:AssumeRoleWithWebIdentity",
				Principal: Principal{
					Federated: "arn:aws:iam::123456789012:oidc-provider/oidc.eks.us-east-2.amazonaws.com/id/EXAMPLED539D4633E53DE1B716D3041",
				},
			},
			want: "c0b4c0c8",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tps := &TrustPolicyStatement{
				Sid:       tt.fields.Sid,
				Effect:    tt.fields.Effect,
				Action:    tt.fields.Action,
				Principal: tt.fields.Principal,
				Condition: tt.fields.Condition,
			}
			if got := tps.Checksum(); got != tt.want {
				t.Errorf("Checksum() = %v, want %v", got, tt.want)
			}
		})
	}
}
