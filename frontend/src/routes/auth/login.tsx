import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { apiClient } from '@/lib/apiClient'
import { useQueryClient } from '@tanstack/react-query'
import { createFileRoute, useNavigate } from '@tanstack/react-router'
import { useState } from 'react'

export const Route = createFileRoute('/auth/login')({
  component: RouteComponent,
})

function RouteComponent() {

  const navigate = useNavigate()
  const queryClient = useQueryClient()

  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()

    const { data, error } = await apiClient.POST("/auth/login", {
      body: {
        email,
        password
      }
    })

    if (error) {
      console.error(error)
      return
    }

    if (data) {
      await queryClient.invalidateQueries({
        queryKey: ['auth/me']
      })
      await navigate({
        to: '/'
      })
    }
  }

  return <div className='flex flex-col h-full items-center justify-center'>
    <h1>Login to oapibase</h1>
    <form className='flex flex-col gap-2' onSubmit={handleSubmit}>
      <label htmlFor='email'>Email</label>
      <Input name="email" type="email" value={email} onChange={(e) => setEmail(e.currentTarget.value)} />
      <label htmlFor='password'>Password</label>
      <Input name="password" type="password" value={password} onChange={(e) => setPassword(e.currentTarget.value)} />
      <Button type="submit">Login</Button>
    </form>
  </div>
}
