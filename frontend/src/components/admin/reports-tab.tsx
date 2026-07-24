"use client";

import { useState, useMemo } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { moderationAPI } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Search } from "lucide-react";
import { toast } from "sonner";
import type { ReportDTO } from "@/types";
import { useLanguage } from "@/contexts/theme";

const NULL_UUID = "00000000-0000-0000-0000-000000000000";

interface ChatMessage {
	id: string;
	sender_id: string;
	content: string;
	created_at: number;
}

export interface GroupedUser {
	reported_user_id: string;
	reported_user_name?: string;
	report_count: number;
	auto_count: number;
	manual_count: number;
	latest_report: ReportDTO;
}

export function ChatViewer({
	reportId,
	reportedUserId,
}: {
	reportId: string;
	reportedUserId: string;
}) {
	const { language, t } = useLanguage();
	const locale = language === "vi" ? "vi-VN" : "en-US";
	const { data: messages = [], isLoading } = useQuery({
		queryKey: ["admin", "report-messages", reportId],
		queryFn: async () => {
			const res = await moderationAPI.getReportMessages(reportId);
			return (res.data as ChatMessage[]) ?? [];
		},
	});

	if (isLoading)
		return <p className="text-sm text-muted-foreground p-4">{t("loading")}</p>;
	if (messages.length === 0)
		return (
			<p className="text-sm text-muted-foreground p-4">
				{t("adminNoEvidenceMessages")}
			</p>
		);

	return (
		<ScrollArea className="h-[380px] pr-2">
			<div className="flex flex-col gap-0.5">
				{messages.map((m) => {
					const isReported = m.sender_id === reportedUserId;
					return (
						<div
							key={m.id}
							className={`flex ${isReported ? "justify-end" : "justify-start"} py-0.5`}
						>
							<div
								className={`max-w-[75%] rounded-lg px-3 py-2 text-sm ${
									isReported
										? "bg-secondary text-secondary-foreground"
										: "bg-muted text-muted-foreground"
								}`}
							>
								<p>{m.content}</p>
								<p className="text-xs opacity-50 mt-0.5">
									{new Date(m.created_at * 1000).toLocaleTimeString(locale)}
								</p>
							</div>
						</div>
					);
				})}
			</div>
		</ScrollArea>
	);
}

export function ReportsTab() {
	const { t } = useLanguage();
	const queryClient = useQueryClient();
	const [search, setSearch] = useState("");
	const [viewing, setViewing] = useState<GroupedUser | null>(null);

	const { data: reports = [], isLoading } = useQuery({
		queryKey: ["admin", "reports"],
		queryFn: async () => {
			const res = await moderationAPI.listReports();
			return (res.data as ReportDTO[]) ?? [];
		},
	});

	const banMutation = useMutation({
		mutationFn: (userId: string) => moderationAPI.banUser(userId),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ["admin", "reports"] });
			queryClient.invalidateQueries({ queryKey: ["admin", "banned-users"] });
			setViewing(null);
			toast.success(t("adminUserBanned"));
		},
		onError: () => toast.error(t("adminFailedToBanUser")),
	});

	const grouped = useMemo<GroupedUser[]>(() => {
		const pending = reports.filter((r) => r.status === "pending");
		const map = new Map<string, GroupedUser>();
		for (const r of pending) {
			const isAuto = r.reporter_id === NULL_UUID;
			const existing = map.get(r.reported_user_id);
			if (!existing) {
				map.set(r.reported_user_id, {
					reported_user_id: r.reported_user_id,
					reported_user_name: r.reported_user_name,
					report_count: 1,
					auto_count: isAuto ? 1 : 0,
					manual_count: isAuto ? 0 : 1,
					latest_report: r,
				});
			} else {
				existing.report_count++;
				if (isAuto) existing.auto_count++;
				else existing.manual_count++;
				if (r.created_at > existing.latest_report.created_at) {
					existing.latest_report = r;
				}
			}
		}
		return Array.from(map.values()).sort((a, b) => b.report_count - a.report_count);
	}, [reports]);

	const filtered = useMemo(() => {
		if (!search.trim()) return grouped;
		const q = search.toLowerCase();
		return grouped.filter(
			(g) =>
				g.reported_user_name?.toLowerCase().includes(q) ||
				g.reported_user_id.toLowerCase().includes(q)
		);
	}, [grouped, search]);

	return (
		<>
			<div className="relative mb-4">
				<Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
				<Input
					placeholder={t("adminSearchNameOrId")}
					value={search}
					onChange={(e) => setSearch(e.target.value)}
					className="pl-9"
				/>
			</div>

			{isLoading && <p className="text-sm text-muted-foreground">{t("loading")}</p>}
			{!isLoading && filtered.length === 0 && (
				<p className="text-sm text-muted-foreground">
					{grouped.length === 0 ? t("adminNoPendingReports") : t("adminNoResults")}
				</p>
			)}

			<div className="grid grid-cols-2 lg:grid-cols-3 gap-3">
				{filtered.map((g) => (
					<button
						key={g.reported_user_id}
						onClick={() => setViewing(g)}
						className="rounded-lg border bg-card p-4 flex flex-col gap-2 shadow-sm text-left hover:bg-accent/50 transition-colors cursor-pointer"
					>
						<div className="flex items-start justify-between gap-2">
							<span className="text-sm font-semibold leading-tight break-all">
								{g.reported_user_name ?? g.reported_user_id.slice(0, 8)}
							</span>
							<Badge className="shrink-0 text-xs">{g.report_count}×</Badge>
						</div>
						<div className="flex gap-1.5 flex-wrap">
							{g.auto_count > 0 && (
								<Badge variant="secondary" className="text-xs px-1.5 py-0">
									{t("adminAutoReports")} {g.auto_count}
								</Badge>
							)}
							{g.manual_count > 0 && (
								<Badge variant="outline" className="text-xs px-1.5 py-0">
									{t("adminManualReports")} {g.manual_count}
								</Badge>
							)}
						</div>
					</button>
				))}
			</div>

			<Dialog open={!!viewing} onOpenChange={(open) => !open && setViewing(null)}>
				<DialogContent className="sm:max-w-lg max-h-[90vh] flex flex-col overflow-hidden">
					<DialogHeader>
						<DialogTitle>
							{viewing?.reported_user_name ?? viewing?.reported_user_id.slice(0, 8)}
						</DialogTitle>
					</DialogHeader>

					{viewing && (
						<>
							<ChatViewer
								reportId={viewing.latest_report.id}
								reportedUserId={viewing.reported_user_id}
							/>
							<div className="pt-2 border-t">
								<Button
									className="w-full"
									onClick={() => banMutation.mutate(viewing.reported_user_id)}
									disabled={banMutation.isPending}
								>
									{t("adminBanUser")}
								</Button>
							</div>
						</>
					)}
				</DialogContent>
			</Dialog>
		</>
	);
}
