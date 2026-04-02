"use client";

import { useState, useMemo } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { moderationAPI } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Search } from "lucide-react";
import { toast } from "sonner";
import { ChatViewer } from "./reports-tab";
import type { BannedUserDTO } from "@/types";

export function BannedUsersTab() {
	const queryClient = useQueryClient();
	const [search, setSearch] = useState("");
	const [viewing, setViewing] = useState<BannedUserDTO | null>(null);

	const { data: users = [], isLoading } = useQuery({
		queryKey: ["admin", "banned-users"],
		queryFn: async () => {
			const res = await moderationAPI.listBannedUsers();
			return (res.data as BannedUserDTO[]) ?? [];
		},
	});

	const unbanMutation = useMutation({
		mutationFn: (userId: string) => moderationAPI.unbanUser(userId),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ["admin", "banned-users"] });
			setViewing(null);
			toast.success("User unbanned");
		},
		onError: () => toast.error("Failed to unban user"),
	});

	const filtered = useMemo(() => {
		if (!search.trim()) return users;
		const q = search.toLowerCase();
		return users.filter(
			(u) =>
				u.name?.toLowerCase().includes(q) ||
				u.email?.toLowerCase().includes(q) ||
				u.id.toLowerCase().includes(q)
		);
	}, [users, search]);

	return (
		<div className="flex flex-col">
			<div className="relative mb-4">
				<Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
				<Input
					placeholder="Search by name, email, or ID..."
					value={search}
					onChange={(e) => setSearch(e.target.value)}
					className="pl-9"
				/>
			</div>

			{isLoading && <p className="text-sm text-muted-foreground">Loading...</p>}
			{!isLoading && filtered.length === 0 && (
				<p className="text-sm text-muted-foreground">
					{users.length === 0 ? "No banned users." : "No results for your search."}
				</p>
			)}

			<div className="grid grid-cols-2 lg:grid-cols-3 gap-3">
				{filtered.map((u) => (
					<button
						key={u.id}
						onClick={() => setViewing(u)}
						className="rounded-lg border bg-card p-4 flex flex-col gap-1.5 shadow-sm text-left hover:bg-accent/50 transition-colors cursor-pointer"
					>
						<span className="text-sm font-semibold leading-tight truncate w-full">
							{u.name ?? u.id.slice(0, 8)}
						</span>
						{u.email && (
							<span className="text-xs text-muted-foreground truncate w-full">
								{u.email}
							</span>
						)}
						<span className="text-xs text-muted-foreground mt-1">
							Banned {new Date(u.created_at * 1000).toLocaleDateString()}
						</span>
					</button>
				))}
			</div>

			<Dialog open={!!viewing} onOpenChange={(open) => !open && setViewing(null)}>
				<DialogContent className="sm:max-w-lg max-h-[90vh] flex flex-col overflow-hidden">
					<DialogHeader>
						<DialogTitle>{viewing?.name ?? viewing?.id.slice(0, 8)}</DialogTitle>
					</DialogHeader>

					{viewing && (
						<>
							{viewing.last_room_id ? (
								<ChatViewer
									roomId={viewing.last_room_id}
									reportedUserId={viewing.id}
								/>
							) : (
								<p className="text-sm text-muted-foreground py-4">
									No chat history available.
								</p>
							)}
							<div className="pt-2 border-t">
								<Button
									variant="outline"
									className="w-full"
									onClick={() => unbanMutation.mutate(viewing.id)}
									disabled={unbanMutation.isPending}
								>
									Unban User
								</Button>
							</div>
						</>
					)}
				</DialogContent>
			</Dialog>
		</div>
	);
}
