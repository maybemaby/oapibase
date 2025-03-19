import { createFileRoute, Link, Outlet, redirect } from '@tanstack/react-router'

export const Route = createFileRoute('/_main-layout/profile')({
  component: RouteComponent,
  beforeLoad(ctx) {
    if (!ctx.context.auth) {
      throw redirect({
        to: '/auth/login',
      })
    }
  },
})

function RouteComponent() {
  return (
    <>
      <nav className='py-2 flex gap-6 border-b-[1px]'>
        <Link to="/profile">Profile</Link>
        <Link to="/profile/feeds">Feeds</Link>
        <Link to="/profile/schedule">Schedule</Link>
      </nav>
      <Outlet />
    </>
  )
}
