// DistribusiChart — histogram distribusi nilai (SVG murni, tanpa library).
// Menampilkan 10 bucket: 0–9, 10–19, ..., 90–100 (nilai 100 masuk bucket
// terakhir). Komponen ini hanya MENAMPILKAN data; nilai dihitung backend
// (kontrak: frontend tidak pernah menghitung nilai).
// Dipakai halaman Analisis utama; bukan halaman tersendiri.

type DistribusiChartProps = {
  /** Nilai akhir peserta (0–100). Nilai di luar rentang dijepit. */
  nilai: number[];
};

const BUCKET_COUNT = 10;

export function DistribusiChart({ nilai }: DistribusiChartProps) {
  if (nilai.length === 0) {
    return (
      <p className="text-sm text-muted">
        Tidak ada data nilai untuk ditampilkan.
      </p>
    );
  }

  const buckets = Array<number>(BUCKET_COUNT).fill(0);
  for (const raw of nilai) {
    const v = Math.min(100, Math.max(0, raw));
    const idx = v === 100 ? BUCKET_COUNT - 1 : Math.floor(v / 10);
    buckets[idx] += 1;
  }
  const maks = Math.max(...buckets, 1);

  // Geometri SVG (viewBox 640x260): area batang y 20..200, label di bawah.
  const W = 640;
  const H = 260;
  const plotTop = 20;
  const plotBottom = 200;
  const plotHeight = plotBottom - plotTop;
  const slot = W / BUCKET_COUNT;
  const barWidth = Math.min(40, slot * 0.6);

  return (
    <div>
      <svg
        viewBox={`0 0 ${W} ${H}`}
        className="w-full"
        role="img"
        aria-label={`Histogram distribusi nilai dari ${nilai.length} peserta dalam 10 bucket`}
      >
        {buckets.map((count, i) => {
          const barHeight = (count / maks) * plotHeight;
          const x = i * slot + (slot - barWidth) / 2;
          const y = plotBottom - barHeight;
          const label = i === BUCKET_COUNT - 1 ? "90–100" : `${i * 10}–${i * 10 + 9}`;
          return (
            <g key={label}>
              <rect
                x={x}
                y={y}
                width={barWidth}
                height={barHeight}
                fill="var(--color-anbk)"
                rx={4}
              />
              {count > 0 && (
                <text
                  x={x + barWidth / 2}
                  y={y - 6}
                  textAnchor="middle"
                  fontSize={13}
                  fontWeight={600}
                  fill="var(--color-ink)"
                >
                  {count}
                </text>
              )}
              <text
                x={i * slot + slot / 2}
                y={plotBottom + 20}
                textAnchor="middle"
                fontSize={12}
                fill="var(--color-inactive)"
              >
                {label}
              </text>
            </g>
          );
        })}
        <line
          x1={0}
          y1={plotBottom}
          x2={W}
          y2={plotBottom}
          stroke="var(--color-border)"
          strokeWidth={1}
        />
      </svg>
      <p className="mt-1 text-xs text-muted">
        {nilai.length} peserta. Bucket terakhir (90–100) termasuk nilai 100.
      </p>
    </div>
  );
}
