// ErrorBoundary — React error boundary (Tim Tambahan 3).
// Menangkap error render di subtree-nya dan menampilkan pesan ramah +
// tombol muat ulang, alih-alih layar putih. Dipakai tim integrasi dengan
// membungkus <App /> atau per-rute, mis. di main.tsx:
//   <ErrorBoundary><App /></ErrorBoundary>
// Migrasi Tim Layout: markup rakitan -> antd Result (status error). Teks tidak berubah.
import { Component, type ErrorInfo, type ReactNode } from "react";
import { Button, Result } from "antd";

interface ErrorBoundaryProps {
  children: ReactNode;
  fallback?: ReactNode;
}

interface ErrorBoundaryState {
  error: Error | null;
}

export default class ErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  state: ErrorBoundaryState = { error: null };

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    return { error };
  }

  componentDidCatch(error: Error, info: ErrorInfo): void {
    // Log untuk debugging; jangan tampilkan stack trace ke pengguna.
    console.error("ErrorBoundary menangkap error:", error, info.componentStack);
  }

  handleReload = (): void => {
    window.location.reload();
  };

  render(): ReactNode {
    if (this.state.error) {
      if (this.props.fallback) return this.props.fallback;
      return (
        <div className="mx-auto max-w-md px-4 py-16">
          <Result
            status="error"
            title="Terjadi kesalahan"
            subTitle="Aplikasi mengalami gangguan saat menampilkan halaman ini. Coba muat ulang halaman. Bila masalah berlanjut, hubungi admin."
            extra={[
              <Button key="reload" onClick={this.handleReload} type="primary">
                Muat Ulang
              </Button>,
              <Button key="home" href="/">
                Kembali ke Beranda
              </Button>,
            ]}
          />
        </div>
      );
    }
    return this.props.children;
  }
}
