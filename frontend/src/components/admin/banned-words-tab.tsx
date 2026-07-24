"use client";

import { useState, useMemo } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { moderationAPI } from "@/lib/api";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "@/components/ui/select";
import { Pencil, Trash2, Check, X } from "lucide-react";
import { toast } from "sonner";
import type { BannedWordDTO } from "@/types";
import { useLanguage } from "@/contexts/theme";

const DEFAULT_CATEGORY = "General";
const NEW_CATEGORY_VALUE = "__new_category__";

interface EditingState {
	id: string;
	word: string;
	category: string;
}

export function BannedWordsTab() {
	const { t } = useLanguage();
	const queryClient = useQueryClient();
	const [newWord, setNewWord] = useState("");
	const [newCategory, setNewCategory] = useState("");
	const [isCreatingCategory, setIsCreatingCategory] = useState(false);
	const [editing, setEditing] = useState<EditingState | null>(null);

	const { data: words = [], isLoading } = useQuery({
		queryKey: ["admin", "words"],
		queryFn: async () => {
			const res = await moderationAPI.listWords();
			return (res.data as BannedWordDTO[]) ?? [];
		},
	});

	// Categories are derived from the groups currently returned by the API.
	const categories = useMemo(
		() => Array.from(new Set(words.map((w) => w.category || DEFAULT_CATEGORY))).sort(),
		[words]
	);
	const selectedCategory = newCategory || categories[0] || DEFAULT_CATEGORY;
	const canAdd =
		newWord.trim().length > 0 &&
		(!isCreatingCategory || newCategory.trim().length > 0);

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
			moderationAPI.addWord(newWord.trim(), selectedCategory.trim() || DEFAULT_CATEGORY),
		onSuccess: () => {
			setNewWord("");
			setNewCategory("");
			setIsCreatingCategory(false);
			queryClient.invalidateQueries({ queryKey: ["admin", "words"] });
			toast.success(t("adminWordAdded"));
		},
		onError: () => toast.error(t("adminFailedToAddWord")),
	});

	const updateMutation = useMutation({
		mutationFn: ({ id, word, category }: EditingState) =>
			moderationAPI.updateWord(id, word.trim(), category.trim() || DEFAULT_CATEGORY),
		onSuccess: () => {
			setEditing(null);
			queryClient.invalidateQueries({ queryKey: ["admin", "words"] });
			toast.success(t("adminWordUpdated"));
		},
		onError: () => toast.error(t("adminFailedToUpdateWord")),
	});

	const deleteMutation = useMutation({
		mutationFn: (id: string) => moderationAPI.deleteWord(id),
		onSuccess: () => queryClient.invalidateQueries({ queryKey: ["admin", "words"] }),
		onError: () => toast.error(t("adminFailedToRemoveWord")),
	});

	const handleAddKeyDown = (e: React.KeyboardEvent) => {
		if (e.key === "Enter" && canAdd) addMutation.mutate();
	};

	return (
		<div className="flex flex-col gap-4">
			{/* Add word form */}
			<div className="flex gap-2 pb-4 border-b">
				<Input
					placeholder={t("adminNewWord")}
					value={newWord}
					onChange={(e) => setNewWord(e.target.value)}
					onKeyDown={handleAddKeyDown}
					className="flex-1"
				/>
				{isCreatingCategory ? (
					<div className="flex w-48 gap-1">
						<Input
							placeholder={t("adminNewCategory")}
							value={newCategory}
							onChange={(e) => setNewCategory(e.target.value)}
							onKeyDown={handleAddKeyDown}
							className="min-w-0"
							autoFocus
						/>
						<Button
							type="button"
							variant="ghost"
							size="icon"
							onClick={() => {
								setNewCategory("");
								setIsCreatingCategory(false);
							}}
							title={t("adminUseExistingCategory")}
							aria-label={t("adminUseExistingCategory")}
						>
							<X size={14} />
						</Button>
					</div>
				) : (
					<Select
						value={selectedCategory}
						onValueChange={(value) => {
							if (value === NEW_CATEGORY_VALUE) {
								setNewCategory("");
								setIsCreatingCategory(true);
								return;
							}
							setNewCategory(value);
						}}
					>
						<SelectTrigger className="w-48">
							<SelectValue placeholder={t("adminSelectCategory")} />
						</SelectTrigger>
						<SelectContent>
							{categories.length === 0 && (
								<SelectItem value={DEFAULT_CATEGORY}>{DEFAULT_CATEGORY}</SelectItem>
							)}
							{categories.map((category) => (
								<SelectItem key={category} value={category}>
									{category}
								</SelectItem>
							))}
							<SelectItem value={NEW_CATEGORY_VALUE}>
								{t("adminNewCategoryAction")}
							</SelectItem>
						</SelectContent>
					</Select>
				)}
				<Button
					onClick={() => addMutation.mutate()}
					disabled={!canAdd || addMutation.isPending}
				>
					{t("adminAdd")}
				</Button>
			</div>

			{isLoading && <p className="text-sm text-muted-foreground">{t("loading")}</p>}
			{!isLoading && words.length === 0 && (
				<p className="text-sm text-muted-foreground">{t("adminNoBannedWords")}</p>
			)}

			{/* Grouped sections */}
			{grouped.map(([category, catWords]) => (
				<div key={category} className="flex flex-col gap-2">
					<h3 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
						{category}
					</h3>
					<div className="flex flex-wrap items-start gap-2">
						{catWords.map((w) =>
							editing?.id === w.id ? (
								/* Inline edit card */
								<div
									key={w.id}
									className="w-40 rounded-md border bg-card px-2 py-1.5 flex flex-col gap-1 shadow-sm"
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
									<Select
										value={editing.category}
										onValueChange={(category) =>
											setEditing({ ...editing, category })
										}
									>
										<SelectTrigger size="sm" className="h-6 w-full px-1 text-xs">
											<SelectValue placeholder={t("adminCategory")} />
										</SelectTrigger>
										<SelectContent>
											{categories.map((category) => (
												<SelectItem key={category} value={category}>
													{category}
												</SelectItem>
											))}
										</SelectContent>
									</Select>
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
									className="w-40 max-w-full rounded-md border bg-card px-2.5 py-1.5 flex items-center gap-1.5 shadow-sm group"
								>
									<span className="min-w-0 flex-1 truncate font-mono text-xs">{w.word}</span>
									<button
										onClick={() =>
											setEditing({
												id: w.id,
												word: w.word,
												category: w.category || DEFAULT_CATEGORY,
											})
										}
										className="opacity-0 group-hover:opacity-100 p-0.5 rounded text-muted-foreground hover:text-foreground transition-opacity cursor-pointer"
										title={t("adminEdit")}
										aria-label={t("adminEdit")}
									>
										<Pencil size={11} />
									</button>
									<button
										onClick={() => deleteMutation.mutate(w.id)}
										disabled={deleteMutation.isPending}
										className="opacity-0 group-hover:opacity-100 p-0.5 rounded text-muted-foreground hover:text-destructive transition-opacity cursor-pointer"
										title={t("adminDelete")}
										aria-label={t("adminDelete")}
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
