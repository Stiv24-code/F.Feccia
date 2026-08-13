export interface LatLng {
  lat?: number | null;
  lng?: number | null;
}

// Distanza in linea d'aria (non un percorso stradale reale) — utile per
// ordinare/filtrare opzioni per vicinanza quando non è disponibile un
// routing calcolato lato backend per quella coppia di punti.
export const haversineKm = (a?: LatLng | null, b?: LatLng | null) => {
  if (a?.lat == null || a?.lng == null || b?.lat == null || b?.lng == null) return null;
  const R = 6371;
  const toRad = (d: number) => (d * Math.PI) / 180;
  const dLat = toRad(b.lat - a.lat);
  const dLng = toRad(b.lng - a.lng);
  const s = Math.sin(dLat / 2) ** 2 + Math.cos(toRad(a.lat)) * Math.cos(toRad(b.lat)) * Math.sin(dLng / 2) ** 2;
  return Math.round(2 * R * Math.asin(Math.sqrt(s)));
};
