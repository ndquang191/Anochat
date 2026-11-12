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
import { Clock3, ListFilter, Save, Settings2, UsersRound } from "lucide-react";

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

	if (isLoading) return <div className="h-72 max-w-4xl animate-pulse rounded-2xl border bg-muted/40" />;
	if (isError || !data) return <div className="rounded-lg border border-destructive/30 bg-destructive/5 p-4 text-sm text-destructive">{t("adminMatchSettingsLoadFailed")}</div>;

	const unchanged = defaultMode === data.default_mode && allowUserChoice === data.allow_user_choice &&
		rematchCooldownSeconds === data.rematch_cooldown_seconds && queueDisplayMode === data.queue_display_mode &&
		queueCountMinimum === data.queue_count_minimum && queueMessageVI === data.queue_message_vi &&
		queueMessageEN === data.queue_message_en;
	return (
		<div className="max-w-4xl space-y-4 pb-4">
			<div className="flex items-start gap-3 rounded-2xl border bg-card p-4 shadow-sm sm:p-5">
				<div className="flex size-10 shrink-0 items-center justify-center rounded-xl bg-primary/10 text-primary">
					<Settings2 className="size-5" aria-hidden="true" />
				</div>
				<div className="min-w-0">
					<h2 className="text-base font-semibold sm:text-lg">{t("adminSettings")}</h2>
					<p className="mt-1 text-sm leading-6 text-muted-foreground">{t("adminMatchSettingsDescription")}</p>
				</div>
			</div>

			<div className="grid gap-4 lg:grid-cols-2">
				<section className="rounded-2xl border bg-card p-4 shadow-sm sm:p-5">
					<div className="mb-5 flex items-center gap-3 border-b pb-4">
						<div className="flex size-9 items-center justify-center rounded-lg bg-blue-500/10 text-blue-600 dark:text-blue-400"><UsersRound className="size-4" aria-hidden="true" /></div>
						<h3 className="font-semibold">{t("matchMode")}</h3>
					</div>
					<div className="space-y-5">
						<div className="space-y-2">
							<Label htmlFor="default-match-mode">{t("defaultMatchMode")}</Label>
							<Select value={defaultMode} onValueChange={(value) => setDefaultMode(value as MatchMode)} disabled={mutation.isPending}>
								<SelectTrigger id="default-match-mode" className="w-full"><SelectValue /></SelectTrigger>
								<SelectContent position="item-aligned"><SelectItem value="mixed">{t("matchModeMixed")}</SelectItem><SelectItem value="opposite_sex">{t("matchModeOpposite")}</SelectItem></SelectContent>
							</Select>
						</div>
						<div className="flex items-center justify-between gap-4 rounded-xl bg-muted/30 p-4 ring-1 ring-border">
							<div><Label htmlFor="allow-user-match-choice">{t("allowUserMatchChoice")}</Label><p className="mt-1 text-sm leading-5 text-muted-foreground">{t("allowUserMatchChoiceDescription")}</p></div>
							<Switch className="shrink-0" id="allow-user-match-choice" checked={allowUserChoice} onCheckedChange={setAllowUserChoice} disabled={mutation.isPending} />
						</div>
					</div>
				</section>

				<section className="rounded-2xl border bg-card p-4 shadow-sm sm:p-5">
					<div className="mb-5 flex items-center gap-3 border-b pb-4">
						<div className="flex size-9 items-center justify-center rounded-lg bg-amber-500/10 text-amber-600 dark:text-amber-400"><Clock3 className="size-4" aria-hidden="true" /></div>
						<h3 className="font-semibold">{t("rematchCooldown")}</h3>
					</div>
					<div className="space-y-2">
						<Label htmlFor="rematch-cooldown">{t("rematchCooldown")}</Label>
						<Select value={String(rematchCooldownSeconds)} onValueChange={(value) => setRematchCooldownSeconds(Number(value))} disabled={mutation.isPending}>
							<SelectTrigger id="rematch-cooldown" className="w-full"><SelectValue /></SelectTrigger>
							<SelectContent position="item-aligned"><SelectItem value="0">{t("disabled")}</SelectItem><SelectItem value="21600">{t("sixHours")}</SelectItem><SelectItem value="43200">{t("twelveHours")}</SelectItem><SelectItem value="86400">{t("twentyFourHours")}</SelectItem><SelectItem value="604800">{t("adminSevenDays")}</SelectItem></SelectContent>
						</Select>
						<p className="text-sm leading-5 text-muted-foreground">{t("rematchCooldownDescription")}</p>
					</div>
				</section>
			</div>

			<section className="rounded-2xl border bg-card p-4 shadow-sm sm:p-5">
				<div className="mb-5 flex items-center gap-3 border-b pb-4">
					<div className="flex size-9 items-center justify-center rounded-lg bg-violet-500/10 text-violet-600 dark:text-violet-400"><ListFilter className="size-4" aria-hidden="true" /></div>
					<h3 className="font-semibold">{t("queueDisplayMode")}</h3>
				</div>
				<div className="grid gap-5 md:grid-cols-2">
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
					<div className="space-y-2 md:col-span-2"><Label htmlFor="queue-message-vi">{t("queueMessageVI")}</Label><Input id="queue-message-vi" maxLength={500} value={queueMessageVI} onChange={(event) => setQueueMessageVI(event.target.value)} /></div>
					<div className="space-y-2 md:col-span-2"><Label htmlFor="queue-message-en">{t("queueMessageEN")}</Label><Input id="queue-message-en" maxLength={500} value={queueMessageEN} onChange={(event) => setQueueMessageEN(event.target.value)} /></div>
				</>}
			</div>
			</section>

			<div className="sticky bottom-0 flex justify-end rounded-2xl border bg-background/90 p-3 shadow-lg backdrop-blur sm:p-4">
				<Button className="w-full gap-2 sm:w-auto sm:min-w-32" onClick={() => mutation.mutate()} disabled={unchanged || mutation.isPending}>
					<Save className="size-4" aria-hidden="true" />{mutation.isPending ? t("saving") : t("saveChanges")}
				</Button>
			</div>
		</div>
	);
}
