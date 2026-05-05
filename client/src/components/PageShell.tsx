interface PageShellProps {
  children: React.ReactNode
  className?: string
  bottomNav?: boolean
}

export default function PageShell({ children, className = '', bottomNav = true }: PageShellProps) {
  return (
    <div className={`min-h-svh bg-white ${className}`}>
      <div className={bottomNav ? 'pb-20' : ''}>
        {children}
      </div>
    </div>
  )
}
