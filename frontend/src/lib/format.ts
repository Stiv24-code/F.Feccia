/**
 * Formatta un importo in euro con separatore migliaia (punto) per locale IT.
 * Es: 2150 → "2.150", 29750 → "29.750", 950 → "950"
 */
export const formatEuro = (value: number | string | null | undefined): string => {
  const num = typeof value === 'number' ? value : Number(value) || 0;
  return num.toLocaleString('it-IT', {
    minimumFractionDigits: 0,
    maximumFractionDigits: 2,
    useGrouping: true,
  });
};
