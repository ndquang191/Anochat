"use client";

import { useEffect, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { adminAPI } from "@/lib/api";
import type { MatchMode, QueueDisplayMode } from "@/types";
import { useLanguage } from "@/contexts/theme";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Input } from "@/components/ui/input";

export function MatchSettingsTab() {
	const { t } = useLanguage();
	const queryClient = useQueryClient();
	const [defaultMode, setDefaultMode] = useState<MatchMode>("mixed");
	const [allowUserChoice, setAllowUserChoice] = useState(false);
	const [rematchCooldownSeconds, setRematchCooldownSeconds] = useState(0);
	const [queueDisplayMode, setQueueDisplayMode] = useState<QueueDisplayMode>("message");
	const [queueCountMinimum, setQueueCountMinimum] = useState(5);
	const [queueMessageVI, setQueueMessageVI] = useState("");
	const [queueMessageEN, setQueueMessageEN] = useState("");
	const { data, isLoading, isError } = useQuery({
		queryKey: ["admin", "match-settings"],
		queryFn: async () => (await adminAPI.getMatchSettings()).data,
	});

	useEffect(() => {
		if (!data) return;
		setDefaultMode(data.default_mode);
		setAllowUserChoice(data.allow_user_choice);
		setRematchCooldownSeconds(data.rematch_cooldown_seconds);
		setQueueDisplayMode(data.queue_display_mode);
		setQueueCountMinimum(data.queue_count_minimum);
		setQueueMessageVI(data.queue_message_vi);
		setQueueMessageEN(data.queue_message_en);
	}, [data]);

	const mutation = useMutation({
		mutationFn: () => adminAPI.updateMatchSettings({
			default_mode: defaultMode, allow_user_choice: allowUserChoice,
			rematch_cooldown_seconds: rematchCooldownSeconds,
			queue_display_mode: queueDisplayMode, queue_count_minimum: queueCountMinimum,
			queue_message_vi: queueMessageVI, queue_message_en: queueMessageEN,
		}),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ["admin", "match-settings"] });
			queryClient.invalidateQueries({ queryKey: ["user-state"] });
			toast.success(t("adminMatchSettingsSaved"));
		},
		onError: () => toast.error(t("adminMatchSettingsSaveFailed")),
	});

	if (isLoading) return <div className="h-40 animate-pulse rounded-xl border bg-muted/40" />;
	if (isError || !data) return <div className="rounded-lg border border-destructive/30 bg-destructive/5 p-4 text-sm text-destructive">{t("adminMatchSettingsLoadFailed")}</div>;

	const unchanged = defaultMode === data.default_mode && allowUserChoice === data.allow_user_choice &&
		rematchCooldownSeconds === data.rematch_cooldown_seconds && queueDisplayMode === data.queue_display_mode &&
		queueCountMinimum === data.queue_count_minimum && queueMessageVI === data.queue_message_vi &&
		queueMessageEN === data.queue_message_en;
	return (
		<div className="max-w-2xl space-y-6 rounded-xl border bg-card p-5">
			<div>
				<h2 className="font-semibold">{t("matchMode")}</h2>
				<p className="mt-1 text-sm text-muted-foreground">{t("adminMatchSettingsDescription")}</p>
			</div>
			<div className="space-y-2">
				<Label htmlFor="default-match-mode">{t("defaultMatchMode")}</Label>
				<Select value={defaultMode} onValueChange={(value) => setDefaultMode(value as MatchMode)} disabled={mutation.isPending}>
					<SelectTrigger id="default-match-mode" className="w-full"><SelectValue /></SelectTrigger>
					<SelectContent position="item-aligned">
						<SelectItem value="mixed">{t("matchModeMixed")}</SelectItem>
						<SelectItem value="opposite_sex">{t("matchModeOpposite")}</SelectItem>
					</SelectContent>
				</Select>
			</div>
			<div className="flex items-center justify-between gap-4 rounded-lg border p-4">
				<div><Label htmlFor="allow-user-match-choice">{t("allowUserMatchChoice")}</Label><p className="mt-1 text-sm text-muted-foreground">{t("allowUserMatchChoiceDescription")}</p></div>
				<Switch id="allow-user-match-choice" checked={allowUserChoice} onCheckedChange={setAllowUserChoice} />
			</div>
			<div className="space-y-2">
				<Label htmlFor="rematch-cooldown">{t("rematchCooldown")}</Label>
				<Select value={String(rematchCooldownSeconds)} onValueChange={(value) => setRematchCooldownSeconds(Number(value))} disabled={mutation.isPending}>
					<SelectTrigger id="rematch-cooldown" className="w-full"><SelectValue /></SelectTrigger>
					<SelectContent position="item-aligned">
						<SelectItem value="0">{t("disabled")}</SelectItem>
						<SelectItem value="21600">{t("sixHours")}</SelectItem>
						<SelectItem value="43200">{t("twelveHours")}</SelectItem>
						<SelectItem value="86400">{t("twentyFourHours")}</SelectItem>
						<SelectItem value="604800">{t("adminSevenDays")}</SelectItem>
					</SelectContent>
				</Select>
				<p className="text-sm text-muted-foreground">{t("rematchCooldownDescription")}</p>
			</div>
			<div className="space-y-4 rounded-lg border p-4">
				<div className="space-y-2">
					<Label htmlFor="queue-display-mode">{t("queueDisplayMode")}</Label>
					<Select value={queueDisplayMode} onValueChange={(value) => setQueueDisplayMode(value as QueueDisplayMode)} disabled={mutation.isPending}>
						<SelectTrigger id="queue-display-mode" className="w-full"><SelectValue /></SelectTrigger>
						<SelectContent position="item-aligned">
							<SelectItem value="hidden">{t("queueDisplayHidden")}</SelectItem>
							<SelectItem value="message">{t("queueDisplayMessage")}</SelectItem>
							<SelectItem value="count">{t("queueDisplayCount")}</SelectItem>
						</SelectContent>
					</Select>
				</div>
				{queueDisplayMode === "count" && <div className="space-y-2">
					<Label htmlFor="queue-count-minimum">{t("queueCountMinimum")}</Label>
					<Input id="queue-count-minimum" type="number" min={1} max={1000} value={queueCountMinimum} onChange={(event) => setQueueCountMinimum(Number(event.target.value))} />
				</div>}
				{queueDisplayMode !== "hidden" && <>
					<div className="space-y-2"><Label htmlFor="queue-message-vi">{t("queueMessageVI")}</Label><Input id="queue-message-vi" maxLength={500} value={queueMessageVI} onChange={(event) => setQueueMessageVI(event.target.value)} /></div>
					<div className="space-y-2"><Label htmlFor="queue-message-en">{t("queueMessageEN")}</Label><Input id="queue-message-en" maxLength={500} value={queueMessageEN} onChange={(event) => setQueueMessageEN(event.target.value)} /></div>
				</>}
			</div>
			<Button onClick={() => mutation.mutate()} disabled={unchanged || mutation.isPending}>{mutation.isPending ? t("saving") : t("saveChanges")}</Button>
		</div>
	);
}
