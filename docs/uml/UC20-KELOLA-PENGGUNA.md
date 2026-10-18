# UC20 — Kelola Pengguna (Akun Staf)

> Spesifikasi kasus penggunaan baru untuk job UML berikutnya (diagram
> Use Case, Activity, dan Sequence). Ditulis dalam **bahasa desain
> (PIM)**: istilah implementasi (nama berkas, pustaka, jalur endpoint,
> SQL, kolom basis data, tipe basis data) SENGAJA tidak dipakai di sini
> — job UML yang memetakan ke implementasi merujuk `docs/KONTRAK.md`
> §8.7 untuk pemetaan teknisnya.
>
> Penanda: **LAMPIRAN** (administrasi akun, bukan siklus ujian — sama
> seperti UC02, UC16, UC17). Satu diagram Activity + satu diagram
> Sequence untuk UC20; elips UC20 ditambahkan ke Diagram Use Case.
> Status implementasi: backend selesai (migrasi 0013); halaman
> antarmuka BELUM ada saat dokumen ini ditulis.

## Ringkasan

Admin mengelola akun pengguna **staf** (bukan siswa): melihat daftar,
menambah akun baru, melihat detail, mengubah nama/surel/role,
mereset kata sandi, serta mengaktifkan/menonaktifkan akun. Akun siswa
tetap dikelola lewat Kelola Master Data (UC02), bukan di sini.
Tidak ada hapus permanen akun: akun yang tidak dipakai lagi
dinonaktifkan saja. Setiap perubahan tercatat di jejak audit.

## Aktor

| Aktor | Peran |
|---|---|
| Admin | Satu-satunya pelaku UC20. Memiliki seluruh izin kelola pengguna. |
| Sistem | Menegakkan validasi, guard admin terakhir, pencatatan jejak audit, dan penghapusan sesi saat reset kata sandi. |

Staf non-admin (proktor, guru, pengawas, penguji UKK) TIDAK dapat
memakai UC20 — sistem menolak mereka. Siswa bukan sasaran UC20.

## Prasyarat

1. Admin sudah masuk (sesi aktif).
2. Admin memiliki izin kelola pengguna (izin `pengguna.*` — lihat
   § Izin di bawah).

## Pascakondisi

- **Berhasil:** akun staf dibuat/diubah/direset/dinonaktifkan sesuai
  aksi, dan satu baris jejak audit tercatat untuk aksi tersebut.
- **Gagal (validasi/guard):** tidak ada perubahan akun, dan tidak ada
  jejak audit untuk aksi yang gagal (kecuali percobaan login — di
  luar UC20).

## Alur utama

### A. Melihat daftar akun staf

1. Admin membuka daftar pengguna.
2. Sistem menampilkan akun **staf saja** (akun siswa tidak tampil),
   terurut menurut nama, dengan kolom: nama pengguna (username),
   nama, surel (email), role, status aktif, wajib ganti kata sandi,
   masuk terakhir, dan tanggal dibuat.
3. Admin dapat menyaring menurut kata kunci (nama / nama pengguna /
   surel), role, dan status aktif.

### B. Menambah akun staf

1. Admin mengisi formulir akun baru: nama pengguna (username), nama,
   surel (opsional), role, dan kata sandi.
2. Sistem memvalidasi isian (lihat § Aturan bisnis R1–R4).
3. Sistem menyimpan akun baru dengan status aktif dan tanda
   "wajib ganti kata sandi".
4. Sistem mencatat jejak audit pembuatan akun.
5. Sistem menampilkan akun baru dalam daftar.

### C. Melihat detail akun staf

1. Admin memilih satu akun staf dari daftar.
2. Sistem menampilkan detail akun tersebut.

### D. Mengubah akun staf

1. Admin mengubah nama, surel, dan/atau role akun staf.
2. Sistem memvalidasi isian (lihat § Aturan bisnis R2, R5–R6).
3. Sistem menyimpan perubahan dan mencatat jejak audit perubahan.

### E. Mereset kata sandi

1. Admin memasukkan kata sandi baru untuk akun staf.
2. Sistem memvalidasi kata sandi baru (lihat § Aturan bisnis R4).
3. Sistem menyimpan kata sandi baru, menandai akun "wajib ganti
   kata sandi", dan **menghapus seluruh sesi** akun tersebut sehingga
   sesi lama tidak dapat dipakai lagi.
4. Sistem mencatat jejak audit reset kata sandi.

### F. Mengaktifkan / menonaktifkan akun staf

1. Admin mengubah status aktif akun staf.
2. Sistem menegakkan guard (lihat § Aturan bisnis R7–R8).
3. Sistem menyimpan status baru dan mencatat jejak audit perubahan
   status.

## Alur alternatif / pengecualian

| # | Kondisi | Perilaku sistem |
|---|---|---|
| ALT-1 | Admin menyaring daftar dengan role siswa | Sistem menolak: akun siswa dikelola lewat halaman Siswa (UC02). |
| ALT-2 | Nama pengguna sudah dipakai (termasuk beda huruf besar/kecil) | Sistem menolak dengan pesan "Nama pengguna sudah dipakai." |
| ALT-3 | Surel sudah dipakai akun lain | Sistem menolak dengan pesan "Surel sudah dipakai." |
| ALT-4 | Surel tidak valid, nama kosong, atau role bukan role staf | Sistem menolak dengan pesan validasi yang menyebut field-nya. |
| ALT-5 | Role yang dipilih adalah siswa | Sistem menolak: akun siswa dikelola lewat halaman Siswa. |
| ALT-6 | Kata sandi baru kurang dari 8 karakter | Sistem menolak: kata sandi minimal 8 karakter. |
| ALT-7 | Admin mencoba mengubah nama pengguna (username) | Nama pengguna TIDAK berubah — field ini tidak dapat diubah. |
| ALT-8 | Admin mengubah role akun admin terakhir yang masih aktif menjadi non-admin | Sistem menolak: minimal 1 admin aktif harus tersisa. |
| ALT-9 | Admin mencoba menonaktifkan akunnya sendiri | Sistem menolak: tidak dapat menonaktifkan akun sendiri. |
| ALT-10 | Admin mencoba menonaktifkan admin terakhir yang masih aktif | Sistem menolak: minimal 1 admin aktif harus tersisa. |
| ALT-11 | Akun yang ditarget tidak ada, atau akun siswa | Sistem menjawab "tidak ditemukan" (tidak membocorkan keberadaan akun siswa — anti-IDOR). |
| ALT-12 | Aktor bukan admin | Sistem menolak (tidak punya izin). |
| ALT-13 | Admin mengaktifkan kembali akun nonaktif | Diizinkan; akun dapat dipakai masuk lagi. |
| ALT-14 | Admin mengubah surel menjadi kosong | Diizinkan: surel dikosongkan (surel opsional; nama pengguna tetap pengenal masuk). |

## Aturan bisnis

- **R1 — Nama pengguna (username):** wajib diisi, tanpa spasi, unik
  tanpa membedakan huruf besar/kecil (karena masuk memakai
  pencocokan tanpa membedakan huruf besar/kecil). Tidak dapat diubah
  setelah akun dibuat.
- **R2 — Surel (email):** opsional; bila diisi harus berformat surel
  yang sah dan unik (tanpa membedakan huruf besar/kecil). Akun wajib
  punya nama pengguna ATAU surel sebagai pengenal masuk.
- **R3 — Role staf:** hanya 5 role staf yang dapat dipilih: admin,
  proktor, guru, pengawas, penguji UKK. Role siswa tidak dapat
  dipilih di sini.
- **R4 — Kata sandi:** minimal 8 karakter, sama seperti kebijakan
  kata sandi pada masuk/ganti kata sandi (UC01). Disimpan dalam
  bentuk hash, tidak pernah dalam bentuk teks jelas, tidak pernah
  tampil di jejak audit.
- **R5 — Wajib ganti kata sandi:** akun baru dan akun yang direset
  kata sandinya selalu ditandai "wajib ganti kata sandi", sehingga
  pemilik akun wajib menggantinya saat masuk pertama (UC01).
- **R6 — Demosi admin:** mengubah role admin menjadi non-admin hanya
  diizinkan bila masih ada admin aktif lain. Admin nonaktif yang
  didemosi tidak dihitung (ia sudah tidak aktif).
- **R7 — Admin terakhir:** menonaktifkan admin terakhir yang masih
  aktif ditolak. Sistem tidak boleh sampai tanpa admin aktif.
- **R8 — Nonaktifkan diri:** admin tidak dapat menonaktifkan akunnya
  sendiri (dicek sebelum guard admin terakhir, agar pesannya jelas).
- **R9 — Tanpa hapus permanen:** tidak ada aksi hapus akun. Akun yang
  tidak dipakai lagi dinonaktifkan; akun nonaktif tidak dapat masuk.
- **R10 — Reset memutuskan sesi:** reset kata sandi menghapus
  seluruh sesi akun tersebut (termasuk sesi admin yang mereset bila
  ia mereset akunnya sendiri — sesi itu berakhir dan ia wajib masuk
  lagi).
- **R11 — Masuk terakhir:** setiap masuk yang sukses memperbarui
  penanda "masuk terakhir" akun tersebut. Akun yang belum pernah
  masuk menampilkan penanda kosong.
- **R12 — Pemisahan tugas (SoD-6):** pengelola pengguna != pengelola
  jejak audit — UC20 tidak menyediakan ubah/hapus jejak audit
  (jejak audit hanya-tambah).

## Izin (untuk pemetaan RBAC di diagram)

| Izin | Aksi yang dilindungi |
|---|---|
| `pengguna.read` | A (daftar), C (detail) |
| `pengguna.create` | B (tambah) |
| `pengguna.update` | D (ubah) |
| `pengguna.reset-password` | E (reset kata sandi) |
| `pengguna.toggle-aktif` | F (aktifkan/nonaktifkan) |

Kelima izin di atas hanya dimiliki role admin. Tidak ada izin
`pengguna.delete` — tidak ada aksi hapus.

## Jejak audit (untuk pemetaan Sequence)

| Aksi | Entitas | Detail yang dicatat |
|---|---|---|
| `pengguna.create` | Pengguna | nama pengguna, nama, surel, role |
| `pengguna.update` | Pengguna | nama, surel, role |
| `pengguna.reset-password` | Pengguna | nama pengguna (tanpa kata sandi) |
| `pengguna.toggle-aktif` | Pengguna | nama pengguna, status aktif baru |

## Objek / data yang tampil (untuk pemetaan Class / Model Data)

Akun Pengguna (staf): nama pengguna, nama, surel, role, status
aktif, wajib ganti kata sandi, masuk terakhir, tanggal dibuat.
Akun siswa memakai objek yang sama tetapi dikelola UC02 dan tidak
pernah tampil di UC20.
