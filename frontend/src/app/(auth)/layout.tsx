export default function AuthLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex min-h-screen items-center justify-center bg-gradient-to-br from-gov-navy-dark via-gov-navy to-gov-slate px-4">
      <div className="w-full max-w-md">{children}</div>
    </div>
  );
}
