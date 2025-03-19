import { useState } from "react";
import { FormGroup } from "./forms";
import { Input } from "./ui/input";
import { Select, SelectTrigger, SelectContent, SelectValue, SelectItem } from "./ui/select";
import { Button } from "./ui/button";
import { useAddFeed } from "@/hooks/feeds";

const urlFeedTypes = ["rss"];

export function FeedForm() {


    const [feedType, setType] = useState('rss');
    const [url, setUrl] = useState('');
    const addFeed = useAddFeed();

    const handleSubmit: React.FormEventHandler<HTMLFormElement> = async (e) => {
        e.preventDefault();
        console.log('submitting', feedType, url);
        await addFeed.mutateAsync({ sourceName: "Feed", sourceType: feedType, url });
    }

    return (
        <form className="flex flex-col gap-4" onSubmit={handleSubmit}>
            <FormGroup label="Name" htmlFor="sourceName">
                <Input id="sourceName" name="sourceName" type="text" />
            </FormGroup>
            <FormGroup label="Type" htmlFor="sourceType">
                <Select onValueChange={(v) => setType(v)} value={feedType}>
                    <SelectTrigger className="w-[200px]">
                        {feedType}
                    </SelectTrigger>
                    <SelectContent>
                        <SelectItem value="rss">RSS</SelectItem>
                        <SelectItem value="podcast">Podcast</SelectItem>
                        <SelectItem value="youtube">YouTube</SelectItem>
                    </SelectContent>
                </Select>
            </FormGroup>
            {urlFeedTypes.includes(feedType) && (
                <FormGroup label="URL" htmlFor="sourceUrl">
                    <Input id="sourceUrl" name="sourceUrl" type="text" value={url} onChange={(e) => setUrl(e.currentTarget.value)} />
                </FormGroup>
            )}
            {feedType === "podcast" && (
                <PodcastSearch />
            )}
            <Button type="submit" disabled={addFeed.isPending}>Add Feed</Button>
        </form>
    )
}

function PodcastSearch() {

    const [search, setSearch] = useState('');

    return (
        <div>
            <Input id="podcastSearch" name="podcastSearch" type="text" value={search} onChange={(e) => setSearch(e.currentTarget.value)} placeholder="Search podcasts by title" />
        </div>
    )
}