/**
 * Geocoding API client for Nominatim (OpenStreetMap)
 *
 * Provides city/country search functionality to convert human-readable
 * locations into latitude/longitude coordinates for household location
 * configuration.
 *
 * Uses Nominatim API: https://nominatim.openstreetmap.org
 * Rate limit: 1 request per second
 * Terms: https://operations.osmfoundation.org/policies/nominatim/
 */

/**
 * Represents a geocoding search result from Nominatim
 */
export interface GeocodingResult {
  /** Full formatted address for display (e.g., "San Francisco, California, United States") */
  displayName: string;
  /** Latitude coordinate as number */
  lat: number;
  /** Longitude coordinate as number */
  lon: number;
  /** City name (may be undefined for some locations) */
  city?: string;
  /** State/province name (may be undefined) */
  state?: string;
  /** Country name (may be undefined) */
  country?: string;
  /** Country code in uppercase (e.g., "US", "JP") */
  countryCode?: string;
}

/**
 * Search for locations by query string using Nominatim API
 *
 * @param query - Search query (e.g., "San Francisco, US" or "Tokyo, Japan")
 * @returns Array of geocoding results, empty array if query too short or on error
 *
 * @example
 * ```typescript
 * const results = await searchLocations('San Francisco');
 * console.log(results[0].lat, results[0].lon);
 * ```
 */
export async function searchLocations(query: string): Promise<GeocodingResult[]> {
  // Validate query length (minimum 3 characters)
  if (!query || query.trim().length < 3) {
    return [];
  }

  try {
    // Build query parameters
    const params = new URLSearchParams({
      q: query.trim(),
      format: 'json',
      addressdetails: '1',
      limit: '5', // Limit to 5 results for better UX
    });

    // Fetch from Nominatim API
    // Note: User-Agent header is REQUIRED by Nominatim Terms of Service
    const response = await fetch(
      `https://nominatim.openstreetmap.org/search?${params.toString()}`,
      {
        headers: {
          'User-Agent': 'HomeHub-Dev/1.0 (Development Environment)', // Required by Nominatim TOS
        },
      }
    );

    if (!response.ok) {
      throw new Error(`Geocoding API error: ${response.status} ${response.statusText}`);
    }

    const data = await response.json();

    // Map Nominatim response to our GeocodingResult interface
    return data.map((item: any) => ({
      displayName: item.display_name,
      // Nominatim returns lat/lon as strings, convert to numbers
      lat: parseFloat(item.lat),
      lon: parseFloat(item.lon),
      // Address components (may be undefined)
      // Some locations use "city", others use "town" or "village"
      city: item.address?.city || item.address?.town || item.address?.village,
      state: item.address?.state,
      country: item.address?.country,
      countryCode: item.address?.country_code?.toUpperCase(),
    }));
  } catch (error) {
    console.error('Geocoding search failed:', error);
    // Return empty array on error (graceful degradation)
    return [];
  }
}

/**
 * Derive timezone from coordinates using TimeAPI.io
 *
 * This uses the free TimeAPI.io service to convert geographic coordinates
 * into accurate IANA timezone identifiers. No API key required.
 *
 * API: https://timeapi.io/api/TimeZone/coordinate?latitude={lat}&longitude={lon}
 *
 * @param lat - Latitude coordinate
 * @param lon - Longitude coordinate
 * @returns IANA timezone string (e.g., "America/Los_Angeles", "Asia/Tokyo")
 *
 * @example
 * ```typescript
 * const tz = await deriveTimezone(37.7749, -122.4194);
 * console.log(tz); // "America/Los_Angeles"
 *
 * const tz2 = await deriveTimezone(35.6762, 139.6503);
 * console.log(tz2); // "Asia/Tokyo"
 * ```
 */
export async function deriveTimezone(lat: number, lon: number): Promise<string> {
  try {
    // Call TimeAPI.io - free timezone lookup API (no API key needed)
    const response = await fetch(
      `https://timeapi.io/api/TimeZone/coordinate?latitude=${lat}&longitude=${lon}`
    );

    if (!response.ok) {
      throw new Error(`TimeAPI returned ${response.status}`);
    }

    const data = await response.json();

    // TimeAPI.io returns { timeZone: "America/Los_Angeles", ... }
    if (data.timeZone) {
      return data.timeZone;
    }

    throw new Error('No timezone in response');
  } catch (error) {
    console.warn('Failed to derive timezone from coordinates, using browser timezone as fallback:', error);
    // Fallback to browser timezone if API call fails
    return Intl.DateTimeFormat().resolvedOptions().timeZone;
  }
}

/**
 * Get timezone from coordinates
 *
 * Wrapper around deriveTimezone() for backwards compatibility.
 *
 * @param lat - Latitude coordinate
 * @param lon - Longitude coordinate
 * @returns Promise resolving to IANA timezone string
 */
export async function getTimezoneFromCoords(
  lat: number,
  lon: number
): Promise<string> {
  return deriveTimezone(lat, lon);
}
