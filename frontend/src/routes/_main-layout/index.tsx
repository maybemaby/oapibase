import { createFileRoute, Link } from "@tanstack/react-router";
import { Button } from "@/components/ui/button";
import { FeedItem } from "@/components/feed-item";
import { apiClient } from "@/lib/apiClient";
import { useId } from "react";

export const Route = createFileRoute("/_main-layout/")({
  component: HomeComponent,
  loader: async () => {
    const feeds = await apiClient.GET("/feed");

    if (feeds.error) {
      return {
        error: {
          message: "Failed to fetch feeds",
        },
      }
    }

    return {
      items: feeds.data,
    };
  },
});

const text = "Lorem ipsum dolor sit amet, consectetur adipiscing elit It can be cumbersome to pass data from parents through to children components, since this means that every component in the hierarchy has to accept parameters and pass them through to children."

function HomeComponent() {

  const id = useId();
  const data = Route.useLoaderData();

  return (
    <div className="py-2 h-full">
      <h1 className="text-2xl my-4">Your Feed</h1>
      <div className="flex flex-col gap-4">
        <FeedItem
          item={{
            source: "Hacker News", title: "Hacker News is a great place to find interesting articles", url: "https://news.ycombinator.com/", previewImage: "https://placehold.co/600x400",
            previewText: text
          }} />
        <FeedItem
          item={{
            source: "Hacker News", title: "Hacker News is a great place to find interesting articles", url: "https://news.ycombinator.com/",
            previewText: text
          }} />

        {data.items?.map(item => (
          <FeedItem key={id} item={item} />
        ))}
      </div>
    </div >
  );
}
