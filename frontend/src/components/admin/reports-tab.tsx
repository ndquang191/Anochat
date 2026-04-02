"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { moderationAPI } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { ScrollArea } from "@/components/ui/scroll-area";
import { toast } from "sonner";
import type { ReportDTO } from "@/types";

const NULL_UUID = "00000000-0000-0000-0000-000000000000";

interface ChatMessage {
	id: string;
	sender_id: string;
	content: string;
	created_at: number;
}

function ChatViewer({ roomId, reportedUserId }: { roomId: string; reportedUserId: string }) {
	const { data: messages = [], isLoading } = useQuery({
		queryKey: ["admin", "room-messages", roomId],
		queryFn: async () => {
			const res = await moderationAPI.getRoomMessages(roomId);
			return (res.data as ChatMessage[]) ?? [];
		},
	});

	if (isLoading) return <p className="text-sm text-muted-foreground p-4">Loading...</p>;
	if (messages.length === 0)
		return <p className="text-sm text-muted-foreground p-4">No messages found (may have been deleted).</p>;

	return (
		<ScrollArea className="h-[400px] pr-2">
			<div className="flex flex-col">
				{messages.map((m: ChatMessage) => {
					const isReported = m.sender_id === reportedUserId;
					return (
						<div
							key={m.id}
							className={`flex ${isReported ? "justify-end" : "justify-start"} py-1`}
						>
							<div
								className={`max-w-[75%] rounded-lg px-3 py-2 text-sm ${
									isReported
										? "bg-red-100 text-red-900 dark:bg-red-950 dark:text-red-100"
										: "bg-muted text-foreground"
								}`}
							>
								<p>{m.content}</p>
								<p className="text-xs opacity-50 mt-0.5">
									{new Date(m.created_at * 1000).toLocaleTimeString()}
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
	const queryClient = useQueryClient();
	const [viewingReport, setViewingReport] = useState<ReportDTO | null>(null);

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
			toast.success("User banned");
		},
		onError: () => toast.error("Failed to ban user"),
	});

	return (
		<>
			<div className="flex flex-col">
				{isLoading && <p className="text-sm text-muted-foreground">Loading...</p>}
				{!isLoading && reports.length === 0 && (
					<p className="text-sm text-muted-foreground">No reports yet.</p>
				)}
				{reports.map((r: ReportDTO) => (
					<div key={r.id} className="flex items-center justify-between py-3 border-b last:border-b-0">
						<div className="flex flex-col min-w-0">
							<span className="text-sm font-medium truncate">
								{r.reported_user_name ?? r.reported_user_id.slice(0, 8)}
							</span>
							<span className="text-xs text-muted-foreground">
								{r.reporter_id === NULL_UUID ? "Auto" : "Manual"} ·{" "}
								{new Date(r.created_at * 1000).toLocaleString()}
							</span>
						</div>
						<div className="flex items-center gap-2 flex-shrink-0">
							<Button
								variant="outline"
								size="sm"
								onClick={() => setViewingReport(r)}
							>
								View
							</Button>
							{r.status === "pending" && (
								<Button
									variant="destructive"
									size="sm"
									onClick={() => banMutation.mutate(r.reported_user_id)}
									disabled={banMutation.isPending}
								>
									Ban
								</Button>
							)}
						</div>
					</div>
				))}
			</div>

			<Dialog open={!!viewingReport} onOpenChange={(open) => !open && setViewingReport(null)}>
				<DialogContent className="sm:max-w-lg max-h-[95vh] flex flex-col overflow-hidden">
					<DialogHeader>
						<DialogTitle>
							Chat — {viewingReport?.reported_user_name ?? viewingReport?.reported_user_id.slice(0, 8)}
						</DialogTitle>
					</DialogHeader>
					{viewingReport && (
						<ChatViewer
							roomId={viewingReport.room_id}
							reportedUserId={viewingReport.reported_user_id}
						/>
					)}
				</DialogContent>
			</Dialog>
		</>
	);
}
