/** Formats a number as Indonesian Rupiah, e.g. 7500 → "Rp 7.500". */
export function rupiah(n: number): string {
  return 'Rp ' + Math.round(n).toLocaleString('id-ID');
}
