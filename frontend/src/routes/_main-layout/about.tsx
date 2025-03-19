import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/_main-layout/about')({
  component: () => <div>Hello /_main-layout/about!</div>,
})
