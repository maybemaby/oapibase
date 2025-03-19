import { createFileRoute, Outlet } from '@tanstack/react-router'

export const Route = createFileRoute('/auth')({
    component: RouteComponent,
    beforeLoad: async ({context}) => {
        
    }
})

function RouteComponent() {
    return (
        <div className='h-full'>
            <Outlet />
        </div>
    )
}
