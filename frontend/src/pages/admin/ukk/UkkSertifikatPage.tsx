import { Link, useParams } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { ArrowLeft, Printer } from "lucide-react";
import apiClient from "../../../lib/api";
import type { UkkSertifikat } from "./types";
import { fmtNilai, fmtTanggal } from "./types";

import { Alert, Button, Card, Descriptions, Breadcrumb } from "antd";
export default function UkkSertifikatPage() {
  const { id = "", siswaId = "" } = useParams();

  const certQuery = useQuery({
    queryKey: ["ukk-sertifikat", id, siswaId],
    queryFn: () => apiClient.get<UkkSertifikat>(`/ukk/${id}/sertifikat/${siswaId}`),
    enabled: id !== "" && siswaId !== "",
  });

  const handlePrint = () => window.print();

  if (certQuery.isLoading) return <p className="p-1 text-muted">Memuat...</p>;
  if (certQuery.isError) {
    return (
      <div className="flex flex-col gap-3">
        <Alert type="error" showIcon message={<>Sertifikat belum diterbitkan atau tidak ditemukan. Sertifikat hanya diterbitkan bagi siswa yang lulus UKK (nilai akhir &ge; 70).</>} />
        <Link
          to={`/adm/ukk/${id}`}
          className="inline-flex min-h-[44px] w-fit items-center gap-1 text-sm font-medium text-primary hover:underline"
        >
          <ArrowLeft size={16} /> Kembali ke Detail UKK
        </Link>
      </div>
    );
  }
  const s = certQuery.data;
  if (!s) return <p className="text-sm text-muted">Sertifikat tidak ditemukan.</p>;

  return (
    <div className="flex flex-col gap-4">
      <div className="print:hidden">
        <Breadcrumb items={[{ title: (<Link to="/adm">Dashboard</Link>) }, { title: (<Link to="/adm/ukk">UKK</Link>) }, { title: "Sertifikat" }]} />
      </div>
      <div className="flex flex-wrap items-center justify-between gap-3 print:hidden">
        <Link
          to={`/adm/ukk/${id}`}
          className="inline-flex min-h-[44px] items-center gap-1 text-sm font-medium text-primary hover:underline"
        >
          <ArrowLeft size={16} /> Kembali ke Detail UKK
        </Link>
        <Button onClick={handlePrint} type="primary">
          <Printer size={16} /> Print
        </Button>
      </div>

      <Card className="mx-auto w-full max-w-3xl print:border-0 print:shadow-none">
        <div className="flex flex-col gap-6 p-8 text-center print:p-4">
          <div>
            <p className="text-sm font-semibold uppercase tracking-widest text-muted">
              Sertifikat Uji Kompetensi Keahlian
            </p>
            <h1 className="mt-2 border-b-2 border-anbk pb-2 text-2xl font-bold text-anbk">
              {s.ukk_nama ?? "Uji Kompetensi Keahlian"}
            </h1>
            <p className="mt-1 text-sm text-muted">
              Nomor: <strong>{s.nomor_sertifikat}</strong>
            </p>
          </div>

          <p className="text-sm text-ink">
            Diberikan kepada:
          </p>
          <div>
            <p className="text-3xl font-bold text-anbk">{s.siswa_nama ?? "-"}</p>
            <p className="mt-1 text-sm text-muted">NISN: {s.siswa_nisn ?? "-"}</p>
            <p className="text-sm text-muted">
              Kompetensi Keahlian: {s.jurusan_nama ?? s.jurusan_kode ?? "-"}
            </p>
          </div>

          <p className="mx-auto max-w-xl text-sm leading-relaxed text-ink">
            Dinyatakan <strong>LULUS</strong> Uji Kompetensi Keahlian dengan predikat{" "}
            <strong>{s.predikat ?? "-"}</strong>.
          </p>

<div className="mx-auto w-full max-w-md">
            <Descriptions
              bordered
              column={1}
              size="small"
              items={[
                { key: "teori", label: "Nilai Teori (40%)", children: fmtNilai(s.nilai_teori) },
                { key: "praktik", label: "Nilai Praktik (60%)", children: fmtNilai(s.nilai_praktik) },
                { key: "akhir", label: "Nilai Akhir", children: fmtNilai(s.nilai_akhir) },
              ]}
            />
          </div>

          <div className="grid grid-cols-1 gap-6 pt-4 text-sm text-ink sm:grid-cols-2">
            <div className="flex flex-col gap-10">
              <span>Penguji Internal</span>
              <span className="font-semibold">{s.penguji_internal || "-"}</span>
            </div>
            <div className="flex flex-col gap-10">
              <span>Penguji Eksternal (DUDIKA)</span>
              <span className="font-semibold">{s.penguji_eksternal || "-"}</span>
            </div>
          </div>

          <p className="text-xs text-muted">
            Diterbitkan: {fmtTanggal(s.diterbitkan_at)}
            {s.ukk_tanggal ? ` &middot; Tanggal UKK: ${fmtTanggal(s.ukk_tanggal)}` : ""}
          </p>
          <p className="text-xs text-muted">
            Dokumen ini adalah data/draf sertifikat dari sistem CBT sekolah. Penerbitan sertifikat
            resmi menjadi kewenangan lembaga berwenang (BNSP / LSP / DUDIKA / sekolah) sesuai
            ketentuan yang berlaku.
          </p>
        </div>
      </Card>
    </div>
  );
}
