import { FeedForm } from '@/components/feed-form'
import { FeedPreview } from '@/components/feed-item';
import { useFeeds } from '@/hooks/feeds'
import { createFileRoute, redirect } from '@tanstack/react-router'

export const Route = createFileRoute('/_main-layout/profile/feeds')({
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

  const feeds = useFeeds();

  return <div>
    <h2 className='my-3'>Add Feed</h2>
    <FeedForm />

    <h2 className='my-3'>Existing Feeds</h2>

    {feeds.data?.map(feed => (
      <FeedPreview key={feed.name} feed={feed} />
    ))}
  </div>
}
