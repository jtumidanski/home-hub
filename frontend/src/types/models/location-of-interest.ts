export interface LocationOfInterestAttributes {
  label: string | null;
  placeName: string;
  latitude: number;
  longitude: number;
  createdAt: string;
  updatedAt: string;
}

export interface LocationOfInterest {
  id: string;
  type: "location-of-interest";
  attributes: LocationOfInterestAttributes;
}

export interface LocationOfInterestCreateAttributes {
  label?: string | null;
  placeName: string;
  latitude: number;
  longitude: number;
}

export interface LocationOfInterestUpdateAttributes {
  label: string | null;
}
