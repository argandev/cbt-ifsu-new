// Package platform — fondasi bersama untuk semua paket internal.
//
// Berisi tipe dan helper yang sebelumnya diduplikasi di tiap paket:
//   - AppError: error terstruktur (code, message, httpStatus)
//   - httpx: penulisan response JSON (envelope sukses/error)
//   - pagination: parsing & penulisan hasil paginasi (2 gaya yang ada)
//
// ATURAN: semua helper di sini menghasilkan output yang IDENTIK dengan
// implementasi aslinya di tiap paket. Jangan ubah bentuk JSON apa pun.
package platform
