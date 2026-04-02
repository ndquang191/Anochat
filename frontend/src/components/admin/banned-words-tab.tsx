"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { moderationAPI } from "@/lib/api";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { toast } from "sonner";
import type { BannedWordDTO } from "@/types";

export function BannedWordsTab() {
	const [newWord, setNewWord] = useState("");
	const queryClient = useQueryClient();

	const { data: words = [], isLoading } = useQuery({
		queryKey: ["admin", "words"],
		queryFn: async () => {
			const res = await moderationAPI.listWords();
			return (res.data as BannedWordDTO[]) ?? [];
		},
	});

	const addMutation = useMutation({
		mutationFn: () => moderationAPI.addWord(newWord.trim()),
		onSuccess: () => {
			setNewWord("");
			queryClient.invalidateQueries({ queryKey: ["admin", "words"] });
			toast.success("Word added");
		},
		onError: () => toast.error("Failed to add word"),
	});

	const deleteMutation = useMutation({
		mutationFn: (id: string) => moderationAPI.deleteWord(id),
		onSuccess: () => queryClient.invalidateQueries({ queryKey: ["admin", "words"] }),
		onError: () => toast.error("Failed to remove word"),
	});

	return (
		<div className="flex flex-col">
			<div className="flex gap-2 pb-4 border-b">
				<Input
					placeholder="Enter word..."
					value={newWord}
					onChange={(e) => setNewWord(e.target.value)}
					onKeyDown={(e) => e.key === "Enter" && newWord.trim() && addMutation.mutate()}
				/>
				<Button
					onClick={() => addMutation.mutate()}
					disabled={!newWord.trim() || addMutation.isPending}
				>
					Add
				</Button>
			</div>
			{isLoading && <p className="text-sm text-muted-foreground pt-4">Loading...</p>}
			{!isLoading && words.length === 0 && (
				<p className="text-sm text-muted-foreground pt-4">No banned words yet.</p>
			)}
			{words.map((w: BannedWordDTO) => (
				<div
					key={w.id}
					className="flex items-center justify-between py-3 border-b last:border-b-0"
				>
					<span className="font-mono text-sm">{w.word}</span>
					<Button
						variant="ghost"
						size="sm"
						onClick={() => deleteMutation.mutate(w.id)}
						disabled={deleteMutation.isPending}
						className="text-muted-foreground hover:text-red-500"
					>
						Remove
					</Button>
				</div>
			))}
		</div>
	);
}
