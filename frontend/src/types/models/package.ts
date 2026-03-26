export interface PackageAttributes {
  trackingNumber: string | null;
  carrier: string;
  label: string | null;
  notes: string | null;
  status: string | null;
  private: boolean;
  estimatedDelivery: string | null;
  actualDelivery: string | null;
  lastPolledAt: string | null;
  archivedAt: string | null;
  isOwner: boolean;
  userId: string;
  trackingEvents?: TrackingEventInline[];
  createdAt: string;
  updatedAt: string;
}

export interface TrackingEventInline {
  timestamp: string;
  status: string;
  description: string;
  location: string | null;
  rawStatus: string | null;
}

export interface Package {
  id: string;
  type: "packages";
  attributes: PackageAttributes;
}

export interface TrackingEventAttributes {
  timestamp: string;
  status: string;
  description: string;
  location: string | null;
  rawStatus: string | null;
}

export interface TrackingEvent {
  id: string;
  type: "trackingEvents";
  attributes: TrackingEventAttributes;
}

export interface PackageSummaryAttributes {
  arrivingTodayCount: number;
  inTransitCount: number;
  exceptionCount: number;
}

export interface PackageSummary {
  id: string;
  type: "packageSummaries";
  attributes: PackageSummaryAttributes;
}

export interface CarrierDetectionAttributes {
  trackingNumber: string;
  detectedCarrier: string;
  confidence: string;
}

export interface CarrierDetection {
  id: string;
  type: "carrierDetections";
  attributes: CarrierDetectionAttributes;
}

export type PackageStatus =
  | "pre_transit"
  | "in_transit"
  | "out_for_delivery"
  | "delivered"
  | "exception"
  | "stale"
  | "archived";

export type Carrier = "usps" | "ups" | "fedex";

export const STATUS_LABELS: Record<string, string> = {
  pre_transit: "Pre-Transit",
  in_transit: "In Transit",
  out_for_delivery: "Out for Delivery",
  delivered: "Delivered",
  exception: "Exception",
  stale: "Stale",
  archived: "Archived",
};

export const CARRIER_LABELS: Record<string, string> = {
  usps: "USPS",
  ups: "UPS",
  fedex: "FedEx",
};
