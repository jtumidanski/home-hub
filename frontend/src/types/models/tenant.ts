export interface TenantAttributes {
  name: string;
  createdAt: string;
  updatedAt: string;
}

export interface Tenant {
  id: string;
  type: "tenants";
  attributes: TenantAttributes;
}
