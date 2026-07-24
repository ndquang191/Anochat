"use client";

import { useDeferredValue, useMemo, useState } from "react";
import { useInfiniteQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { moderationAPI } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Search, Scale } from "lucide-react";
import { toast } from "sonner";
import { ChatViewer } from "./reports-tab";
import type { BannedUserDTO, BannedUserPageDTO } from "@/types";
import { useLanguage } from "@/contexts/theme";

export function BannedUsersTab() {
	const { language, t } = useLanguage();
	const locale = language === "vi" ? "vi-VN" : "en-US";
	const queryClient = useQueryClient();
	const [search, setSearch] = useState("");
	const [viewing, setViewing] = useState<BannedUserDTO | null>(null);
	const deferredSearch = useDeferredValue(search.trim());

	const {
		data,
		isLoading,
		fetchNextPage,
		hasNextPage,
		isFetchingNextPage,
	} = useInfiniteQuery({
		queryKey: ["admin", "banned-users", deferredSearch],
		initialPageParam: undefined as string | undefined,
		queryFn: async ({ pageParam }) => {
			const res = await moderationAPI.listBannedUsers({
				before: pageParam,
				query: deferredSearch || undefined,
			});
			return res.data ?? { users: [], has_more: false, total: 0 };
		},
		getNextPageParam: (lastPage: BannedUserPageDTO) =>
			lastPage.has_more ? lastPage.next_cursor : undefined,
	});
	const users = useMemo(
		() => data?.pages.flatMap((page) => page.users) ?? [],
		[data]
	);
	const total = data?.pages[0]?.total ?? 0;

	const unbanMutation = useMutation({
		mutationFn: (userId: string) => moderationAPI.unbanUser(userId),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ["admin", "banned-users"] });
			setViewing(null);
			toast.success(t("adminUserUnbanned"));
		},
		onError: () => toast.error(t("adminFailedToUnbanUser")),
	});

	return (
		<div className="flex flex-col">
			<div className="mb-4 flex items-center gap-3">
				<div className="relative max-w-md flex-1">
					<Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
					<Input
						placeholder={t("adminSearchNameEmailId")}
						value={search}
						onChange={(e) => setSearch(e.target.value)}
						className="pl-9"
					/>
				</div>
				<span className="shrink-0 text-xs text-muted-foreground">
					{t("adminUsersCount", {
						count: total.toLocaleString(locale),
					})}
				</span>
			</div>

			{isLoading && <p className="text-sm text-muted-foreground">{t("loading")}</p>}
			{!isLoading && users.length === 0 && (
				<p className="text-sm text-muted-foreground">
					{deferredSearch ? t("adminNoResults") : t("adminNoBannedUsers")}
				</p>
			)}

			{users.length > 0 && (
				<div className="hidden grid-cols-12 gap-3 border-b px-3 pb-2 text-[11px] font-medium uppercase tracking-wide text-muted-foreground md:grid">
					<span className="col-span-5">{t("adminUserColumn")}</span>
					<span className="col-span-3">{t("adminStatus")}</span>
					<span className="col-span-2">{t("adminHistory")}</span>
					<span className="col-span-2 text-right">{t("adminBannedOn")}</span>
				</div>
			)}

			<div className={`divide-y overflow-hidden rounded-lg border bg-card ${users.length === 0 ? "hidden" : ""}`}>
				{users.map((u) => (
					<button
						key={u.id}
						onClick={() => setViewing(u)}
						className={`grid w-full grid-cols-2 items-center gap-3 px-3 py-3 text-left transition-colors cursor-pointer md:grid-cols-12 ${
							u.review_requested
								? "bg-amber-500/10 hover:bg-amber-500/15"
								: "hover:bg-accent/50"
						}`}
					>
						<span className="col-span-2 min-w-0 md:col-span-5">
							<span className="flex min-w-0 items-baseline gap-2">
								<span className="truncate text-sm font-semibold">
									{u.name ?? u.id.slice(0, 8)}
								</span>
								{u.email && (
									<span className="truncate text-xs text-muted-foreground">
										{u.email}
									</span>
								)}
							</span>
							<span className="block truncate font-mono text-[10px] text-muted-foreground/70">
								{u.id}
							</span>
						</span>

						<span className="md:col-span-3">
							{u.review_requested ? (
								<span className="inline-flex items-center gap-1 rounded-full bg-amber-500/15 px-2 py-1 text-[11px] font-medium text-amber-700 dark:text-amber-300">
									<Scale className="h-3 w-3" />
									{t("adminReviewRequested")}
								</span>
							) : (
								<span className="text-xs text-muted-foreground">
									{t("adminNoReviewRequest")}
								</span>
							)}
						</span>

						<span className="text-xs md:col-span-2">
							<span className="font-medium">{u.ban_count}</span>
							<span className="text-muted-foreground"> {t("adminBans")}</span>
							<span className="mx-1.5 text-muted-foreground/50">·</span>
							<span className="font-medium">{u.review_request_count}</span>
							<span className="text-muted-foreground"> {t("adminReviews")}</span>
						</span>

						<span className="col-span-2 text-xs text-muted-foreground md:col-span-2 md:text-right">
							{new Date(
								(u.banned_at ?? u.created_at) * 1000
							).toLocaleDateString(locale)}
						</span>
					</button>
				))}
			</div>
			{hasNextPage && (
				<Button
					variant="outline"
					className="mt-4 w-full"
					onClick={() => fetchNextPage()}
					disabled={isFetchingNextPage}
				>
					{isFetchingNextPage ? t("loading") : t("loadMore")}
				</Button>
			)}

			<Dialog open={!!viewing} onOpenChange={(open) => !open && setViewing(null)}>
				<DialogContent className="sm:max-w-lg max-h-[90vh] flex flex-col overflow-hidden">
					<DialogHeader>
						<DialogTitle>{viewing?.name ?? viewing?.id.slice(0, 8)}</DialogTitle>
					</DialogHeader>

					{viewing && (
						<>
							{viewing.last_report_id ? (
								<ChatViewer
									reportId={viewing.last_report_id}
									reportedUserId={viewing.id}
								/>
							) : (
								<p className="text-sm text-muted-foreground py-4">
									{t("adminNoChatHistory")}
								</p>
							)}
							<div className="pt-2 border-t">
								<Button
									variant="outline"
									className="w-full"
									onClick={() => unbanMutation.mutate(viewing.id)}
									disabled={unbanMutation.isPending}
								>
									{t("adminUnbanUser")}
								</Button>
							</div>
						</>
					)}
				</DialogContent>
			</Dialog>
		</div>
	);
}
