import { BaseService } from "./base";
import type {
  Package,
  PackageSummary,
  CarrierDetection,
} from "@/types/models/package";

class PackageService extends BaseService {
  constructor() {
    super("/packages");
  }

  getPackages(tenant: { id: string }, params?: string) {
    const path = params ? `/packages?${params}` : "/packages";
    return this.getList<Package>(tenant, path);
  }

  getPackage(tenant: { id: string }, id: string) {
    return this.getOne<Package>(tenant, `/packages/${id}`);
  }

  createPackage(
    tenant: { id: string },
    attrs: {
      trackingNumber: string;
      carrier: string;
      label?: string;
      notes?: string;
      private?: boolean;
    }
  ) {
    return this.create<Package>(tenant, "/packages", {
      data: {
        type: "packages",
        attributes: attrs,
      },
    });
  }

  updatePackage(
    tenant: { id: string },
    id: string,
    attrs: {
      label?: string;
      notes?: string;
      carrier?: string;
      private?: boolean;
    }
  ) {
    return this.update<Package>(tenant, `/packages/${id}`, {
      data: {
        type: "packages",
        id,
        attributes: attrs,
      },
    });
  }

  deletePackage(tenant: { id: string }, id: string) {
    return this.remove(tenant, `/packages/${id}`);
  }

  archivePackage(tenant: { id: string }, id: string) {
    return this.create<Package>(tenant, `/packages/${id}/archive`, {});
  }

  unarchivePackage(tenant: { id: string }, id: string) {
    return this.create<Package>(tenant, `/packages/${id}/unarchive`, {});
  }

  refreshPackage(tenant: { id: string }, id: string) {
    return this.create<Package>(tenant, `/packages/${id}/refresh`, {});
  }

  getSummary(tenant: { id: string }) {
    return this.getOne<PackageSummary>(tenant, "/packages/summary");
  }

  detectCarrier(tenant: { id: string }, trackingNumber: string) {
    return this.getOne<CarrierDetection>(
      tenant,
      `/packages/carriers/detect?trackingNumber=${encodeURIComponent(trackingNumber)}`
    );
  }
}

export const packageService = new PackageService();
