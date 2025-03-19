import { Button } from '@/components/ui/button'
import { apiClient } from '@/lib/apiClient'
import { useMutation } from '@tanstack/react-query'
import { createFileRoute, redirect } from '@tanstack/react-router'

export const Route = createFileRoute('/_main-layout/profile/schedule')({
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

  const sampleM = useMutation({
    mutationFn: async () => {
      const res = await apiClient.POST("/windows", {
        body: {
          days: [0, 1, 2, 3, 4, 5, 6],
          start: "00:00:00",
          end: "23:59:00",
          name: "Sample Window"
        }
      })

      if (res.error) {
        throw new Error("Failed to add sample window")
      }

      return res.data
    }
  })


  return <div className='py-4'>
    <h1>Schedule</h1>

    <Button onClick={() => sampleM.mutate()}>Add Sample Window</Button>

  </div>
}
