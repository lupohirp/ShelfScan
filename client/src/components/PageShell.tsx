interface PageShellProps {
  children: React.ReactNode
  className?: string
  bottomNav?: boolean
}

export default function PageShell({ children, className = '', bottomNav = true }: PageShellProps) {
  return (
    <div className={`min-h-svh ${className}`}>
      <div className={bottomNav ? 'pb-24' : ''}>
        {children}
      </div>
    </div>
  )
}
