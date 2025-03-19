import { useDeleteFeed } from "@/hooks/feeds"
import { components } from "@/lib/api"
import { Button } from "./ui/button";


export function FeedItem({ item }: { item: components["schemas"]["feedItem"] }) {
    return (
        <div className="flex gap-4">
            {item.previewImage && (
                <img src={item.previewImage} className="w-24 h-24 object-cover" />
            )}

            <div className="max-w-[32ch] md:max-w-[52ch]">
                <a className="text-lg" href={item.url} target="_blank">{item.title}</a>
                {item.previewText && (
                    <p className="line-clamp-2 text-sm">{item.previewText}</p>
                )}
            </div>
        </div>
    )
}

export function FeedPreview({ feed }: { feed: components["schemas"]["getFeed"] }) {

    const deleteFeed = useDeleteFeed(feed.id);

    return (
        <div className="border p-2 overflow-hidden">
            <p className="text-lg">{feed.name}</p>
            <p>{feed.feedType}</p>
            <pre>{feed.url}</pre>
            <Button size={"sm"} variant={"destructive"} onClick={() => deleteFeed.mutate()}>Remove Feed</Button>
        </div>
    )
}