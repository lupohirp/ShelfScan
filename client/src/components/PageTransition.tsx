import { motion } from 'framer-motion'
import { useLocation } from 'react-router-dom'

interface PageTransitionProps {
  children: React.ReactNode
}

const variants = {
  initial: {
    opacity: 0,
    x: 20,
  },
  enter: {
    opacity: 1,
    x: 0,
    transition: {
      duration: 0.3,
      ease: [0.16, 1, 0.3, 1], // easeOutQuart
    },
  },
  exit: {
    opacity: 0,
    x: -20,
    transition: {
      duration: 0.2,
      ease: [0.7, 0, 0.84, 0], // easeInCirc
    },
  },
}

export default function PageTransition({ children }: PageTransitionProps) {
  const { pathname } = useLocation()

  return (
    <motion.div
      key={pathname}
      initial="initial"
      animate="enter"
      exit="exit"
      variants={variants}
      className="w-full min-h-svh"
    >
      {children}
    </motion.div>
  )
}
