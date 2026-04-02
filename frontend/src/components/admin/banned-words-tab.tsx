"use client";

import { useState, useMemo } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { moderationAPI } from "@/lib/api";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Pencil, Trash2, Check, X } from "lucide-react";
import { toast } from "sonner";
import type { BannedWordDTO } from "@/types";

const DEFAULT_CATEGORY = "General";

interface EditingState {
	id: string;
	word: string;
	category: string;
}

export function BannedWordsTab() {
	const queryClient = useQueryClient();
	const [newWord, setNewWord] = useState("");
	const [newCategory, setNewCategory] = useState("");
	const [editing, setEditing] = useState<EditingState | null>(null);

	const { data: words = [], isLoading } = useQuery({
		queryKey: ["admin", "words"],
		queryFn: async () => {
			const res = await moderationAPI.listWords();
			return (res.data as BannedWordDTO[]) ?? [];
		},
	});

	// Collect existing category names for datalist autocomplete
	const categories = useMemo(
		() => Array.from(new Set(words.map((w) => w.category || DEFAULT_CATEGORY))).sort(),
		[words]
	);

	// Group words by category
	const grouped = useMemo(() => {
		const map = new Map<string, BannedWordDTO[]>();
		for (const w of words) {
			const cat = w.category || DEFAULT_CATEGORY;
			if (!map.has(cat)) map.set(cat, []);
			map.get(cat)!.push(w);
		}
		return Array.from(map.entries()).sort(([a], [b]) => a.localeCompare(b));
	}, [words]);

	const addMutation = useMutation({
		mutationFn: () =>
			moderationAPI.addWord(newWord.trim(), newCategory.trim() || DEFAULT_CATEGORY),
		onSuccess: () => {
			setNewWord("");
			queryClient.invalidateQueries({ queryKey: ["admin", "words"] });
			toast.success("Word added");
		},
		onError: () => toast.error("Failed to add word"),
	});

	const updateMutation = useMutation({
		mutationFn: ({ id, word, category }: EditingState) =>
			moderationAPI.updateWord(id, word.trim(), category.trim() || DEFAULT_CATEGORY),
		onSuccess: () => {
			setEditing(null);
			queryClient.invalidateQueries({ queryKey: ["admin", "words"] });
			toast.success("Word updated");
		},
		onError: () => toast.error("Failed to update word"),
	});

	const deleteMutation = useMutation({
		mutationFn: (id: string) => moderationAPI.deleteWord(id),
		onSuccess: () => queryClient.invalidateQueries({ queryKey: ["admin", "words"] }),
		onError: () => toast.error("Failed to remove word"),
	});

	const handleAddKeyDown = (e: React.KeyboardEvent) => {
		if (e.key === "Enter" && newWord.trim()) addMutation.mutate();
	};

	return (
		<div className="flex flex-col gap-4">
			{/* Add word form */}
			<div className="flex gap-2 pb-4 border-b">
				<Input
					placeholder="New word..."
					value={newWord}
					onChange={(e) => setNewWord(e.target.value)}
					onKeyDown={handleAddKeyDown}
					className="flex-1"
				/>
				<Input
					placeholder="Category (optional)"
					value={newCategory}
					onChange={(e) => setNewCategory(e.target.value)}
					onKeyDown={handleAddKeyDown}
					list="category-list"
					className="w-40"
				/>
				<datalist id="category-list">
					{categories.map((c) => (
						<option key={c} value={c} />
					))}
				</datalist>
				<Button
					onClick={() => addMutation.mutate()}
					disabled={!newWord.trim() || addMutation.isPending}
				>
					Add
				</Button>
			</div>

			{isLoading && <p className="text-sm text-muted-foreground">Loading...</p>}
			{!isLoading && words.length === 0 && (
				<p className="text-sm text-muted-foreground">No banned words yet.</p>
			)}

			{/* Grouped sections */}
			{grouped.map(([category, catWords]) => (
				<div key={category} className="flex flex-col gap-2">
					<h3 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
						{category}
					</h3>
					<div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-2">
						{catWords.map((w) =>
							editing?.id === w.id ? (
								/* Inline edit card */
								<div
									key={w.id}
									className="rounded-md border bg-card px-2 py-1.5 flex flex-col gap-1 shadow-sm"
								>
									<Input
										value={editing.word}
										onChange={(e) =>
											setEditing({ ...editing, word: e.target.value })
										}
										onKeyDown={(e) => {
											if (e.key === "Enter") updateMutation.mutate(editing);
											if (e.key === "Escape") setEditing(null);
										}}
										className="h-6 text-xs px-1"
										autoFocus
									/>
									<Input
										value={editing.category}
										onChange={(e) =>
											setEditing({ ...editing, category: e.target.value })
										}
										list="category-list"
										placeholder="Category"
										className="h-6 text-xs px-1"
									/>
									<div className="flex gap-1 justify-end">
										<button
											onClick={() => updateMutation.mutate(editing)}
											disabled={updateMutation.isPending}
											className="p-0.5 rounded text-muted-foreground hover:text-foreground cursor-pointer"
										>
											<Check size={12} />
										</button>
										<button
											onClick={() => setEditing(null)}
											className="p-0.5 rounded text-muted-foreground hover:text-foreground cursor-pointer"
										>
											<X size={12} />
										</button>
									</div>
								</div>
							) : (
								/* Normal word card */
								<div
									key={w.id}
									className="rounded-md border bg-card px-2.5 py-1.5 flex items-center gap-1.5 shadow-sm group"
								>
									<span className="font-mono text-xs flex-1 truncate">{w.word}</span>
									<button
										onClick={() =>
											setEditing({
												id: w.id,
												word: w.word,
												category: w.category || DEFAULT_CATEGORY,
											})
										}
										className="opacity-0 group-hover:opacity-100 p-0.5 rounded text-muted-foreground hover:text-foreground transition-opacity cursor-pointer"
										title="Edit"
									>
										<Pencil size={11} />
									</button>
									<button
										onClick={() => deleteMutation.mutate(w.id)}
										disabled={deleteMutation.isPending}
										className="opacity-0 group-hover:opacity-100 p-0.5 rounded text-muted-foreground hover:text-destructive transition-opacity cursor-pointer"
										title="Delete"
									>
										<Trash2 size={11} />
									</button>
								</div>
							)
						)}
					</div>
				</div>
			))}
		</div>
	);
}
