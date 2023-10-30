package v1alpha1

import "testing"

func TestTrustPolicyStatement_Id(t *testing.T) {
	type fields struct {
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
				Effect: "Allow",
				Action: "sts:AssumeRoleWithWebIdentity",
				Principal: Principal{
					Federated: "arn:aws:iam::123456789012:oidc-provider/oidc.eks.us-east-2.amazonaws.com/id/EXAMPLED539D4633E53DE1B716D3041",
				},
			},
			want: "AllowStsAssumeRoleWithWebIdentityef522ae8",
		},
		{
			name: "test2: with conditions",
			fields: fields{
				Effect: "Allow",
				Action: "sts:AssumeRoleWithWebIdentity",
				Principal: Principal{
					Federated: "arn:aws:iam::123456789012:oidc-provider/oidc.eks.us-east-2.amazonaws.com/id/EXAMPLED539D4633E53DE1B716D3041",
				},
				Condition: &Condition{
					StringEquals: map[string]string{
						"oidc.eks.us-east-2.amazonaws.com/id/EXAMPLED539D4633E53DE1B716D3041:sub": "system:serviceaccount:my-namespace:my-serviceaccount",
					},
				},
			},
			want: "AllowStsAssumeRoleWithWebIdentityef522ae8d16a3945",
		},
		{
			name: "test3",
			fields: fields{
				Effect: "Allow",
				Action: "sts:AssumeRole",
				Principal: Principal{
					AWS: []string{
						"arn:aws:iam::123456789012:root",
					},
				},
			},
			want: "AllowStsAssumeRole21d512f8",
		},
		{
			name: "test4 - any service account",
			fields: fields{
				Effect: "Allow",
				Action: "sts:AssumeRoleWithWebIdentity",
				Principal: Principal{
					Federated: "arn:aws:iam::123456789012:oidc-provider/oidc.eks.us-east-2.amazonaws.com/id/EXAMPLED539D4633E53DE1B716D3041",
				},
				Condition: &Condition{
					StringLike: map[string]string{
						"oidc.eks.us-east-2.amazonaws.com/id/EXAMPLED539D4633E53DE1B716D3041:sub": "system:serviceaccount:*:*",
					},
				},
			},
			want: "AllowStsAssumeRoleWithWebIdentityef522ae8Any",
		},
		{
			name: "test5 - Not any service account",
			fields: fields{
				Effect: "Allow",
				Action: "sts:AssumeRoleWithWebIdentity",
				Principal: Principal{
					Federated: "arn:aws:iam::123456789012:oidc-provider/oidc.eks.us-east-2.amazonaws.com/id/EXAMPLED539D4633E53DE1B716D3041",
				},
				Condition: &Condition{
					StringLike: map[string]string{
						"oidc.eks.us-east-2.amazonaws.com/id/EXAMPLED539D4633E53DE1B716D3041:sub": "system:serviceaccount:*:default",
					},
				},
			},
			want: "AllowStsAssumeRoleWithWebIdentityef522ae8a57730a3",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tps := &TrustPolicyStatement{
				Effect:    tt.fields.Effect,
				Action:    tt.fields.Action,
				Principal: tt.fields.Principal,
				Condition: tt.fields.Condition,
			}
			if got := tps.Id(); got != tt.want {
				t.Errorf("Id() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrustPolicyStatement_HasCondition(t *testing.T) {
	type fields struct {
		Effect    Effect
		Action    string
		Principal Principal
		Condition *Condition
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "test1",
			fields: fields{
				Effect: "Allow",
				Action: "sts:AssumeRoleWithWebIdentity",
				Principal: Principal{
					Federated: "arn:aws:iam::123456789012:oidc-provider/oidc.eks.us-east-2.amazonaws.com/id/EXAMPLED539D4633E53DE1B716D3041",
				},
			},
			want: false,
		},
		{
			name: "test2: with conditions",
			fields: fields{
				Effect: "Allow",
				Action: "sts:AssumeRoleWithWebIdentity",
				Principal: Principal{
					Federated: "arn:aws:iam::123456789012:oidc-provider/oidc.eks.us-east-2.amazonaws.com/id/EXAMPLED539D4633E53DE1B716D3041",
				},
				Condition: &Condition{
					StringEquals: map[string]string{
						"oidc.eks.us-east-2.amazonaws.com/id/EXAMPLED539D4633E53DE1B716D3041:sub": "system:serviceaccount:my-namespace:my-serviceaccount",
					},
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tps := &TrustPolicyStatement{
				Effect:    tt.fields.Effect,
				Action:    tt.fields.Action,
				Principal: tt.fields.Principal,
				Condition: tt.fields.Condition,
			}
			if got := tps.HasCondition(); got != tt.want {
				t.Errorf("HasCondition() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrustPolicyStatement_ConditionChecksum(t *testing.T) {
	type fields struct {
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
			name: "test1 - no condition empty string",
			fields: fields{
				Effect: "Allow",
				Action: "sts:AssumeRoleWithWebIdentity",
				Principal: Principal{
					Federated: "arn:aws:iam::123456789012:oidc-provider/oidc.eks.us-east-2.amazonaws.com/id/EXAMPLED539D4633E53DE1B716D3041",
				},
			},
			want: "",
		},
		{
			name: "test2: with conditions",
			fields: fields{
				Effect: "Allow",
				Action: "sts:AssumeRoleWithWebIdentity",
				Principal: Principal{
					Federated: "arn:aws:iam::123456789012:oidc-provider/oidc.eks.us-east-2.amazonaws.com/id/EXAMPLED539D4633E53DE1B716D3041",
				},
				Condition: &Condition{
					StringEquals: map[string]string{
						"oidc.eks.us-east-2.amazonaws.com/id/EXAMPLED539D4633E53DE1B716D3041:sub": "system:serviceaccount:my-namespace:my-serviceaccount",
					},
				},
			},
			want: "d16a3945",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tps := &TrustPolicyStatement{
				Effect:    tt.fields.Effect,
				Action:    tt.fields.Action,
				Principal: tt.fields.Principal,
				Condition: tt.fields.Condition,
			}
			if got := tps.ConditionChecksum(); got != tt.want {
				t.Errorf("ConditionChecksum() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrustPolicyStatement_IsConditionAnyServiceAccount(t *testing.T) {
	type fields struct {
		Effect    Effect
		Action    string
		Principal Principal
		Condition *Condition
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "test1 - no condition",
			fields: fields{
				Effect: "Allow",
				Action: "sts:AssumeRoleWithWebIdentity",
				Principal: Principal{
					Federated: "arn:aws:iam::123456789012:oidc-provider/oidc.eks.us-east-2.amazonaws.com/id/EXAMPLED539D4633E53DE1B716D3041",
				},
			},
			want: false,
		},
		{
			name: "test2: with conditions - any service account",
			fields: fields{
				Effect: "Allow",
				Action: "sts:AssumeRoleWithWebIdentity",
				Principal: Principal{
					Federated: "arn:aws:iam::123456789012:oidc-provider/oidc.eks.us-east-2.amazonaws.com/id/EXAMPLED539D4633E53DE1B716D3041",
				},
				Condition: &Condition{
					StringLike: map[string]string{
						"oidc.eks.us-east-2.amazonaws.com/id/EXAMPLED539D4633E53DE1B716D3041:sub": "system:serviceaccount:*:*",
					},
				},
			},
			want: true,
		},
		{
			name: "test3: with conditions - any namespace but not any service account",
			fields: fields{
				Effect: "Allow",
				Action: "sts:AssumeRoleWithWebIdentity",
				Principal: Principal{
					Federated: "arn:aws:iam::123456789012:oidc-provider/oidc.eks.us-east-2.amazonaws.com/id/EXAMPLED539D4633E53DE1B716D3041",
				},
				Condition: &Condition{
					StringLike: map[string]string{
						"oidc.eks.us-east-2.amazonaws.com/id/EXAMPLED539D4633E53DE1B716D3041:sub": "system:serviceaccount:*:default",
					},
				},
			},
			want: false,
		},
		{
			name: "test3: with conditions - not like condition",
			fields: fields{
				Effect: "Allow",
				Action: "sts:AssumeRoleWithWebIdentity",
				Principal: Principal{
					Federated: "arn:aws:iam::123456789012:oidc-provider/oidc.eks.us-east-2.amazonaws.com/id/EXAMPLED539D4633E53DE1B716D3041",
				},
				Condition: &Condition{
					StringEquals: map[string]string{
						"oidc.eks.us-east-2.amazonaws.com/id/EXAMPLED539D4633E53DE1B716D3041:sub": "system:serviceaccount:my-ns:default",
					},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tps := &TrustPolicyStatement{
				Effect:    tt.fields.Effect,
				Action:    tt.fields.Action,
				Principal: tt.fields.Principal,
				Condition: tt.fields.Condition,
			}
			if got := tps.IsConditionAnyServiceAccount(); got != tt.want {
				t.Errorf("IsConditionAnyServiceAccount() = %v, want %v", got, tt.want)
			}
		})
	}
}
