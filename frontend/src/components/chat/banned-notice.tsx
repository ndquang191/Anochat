"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { ShieldAlert } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { useLanguage } from "@/contexts/theme";
import { USER_STATE_KEY } from "@/hooks/queries/use-user-state";
import { userAPI } from "@/lib/api";

interface BannedNoticeProps {
	reviewRequested: boolean;
}

export function BannedNotice({ reviewRequested }: BannedNoticeProps) {
	const { t } = useLanguage();
	const queryClient = useQueryClient();
	const reviewMutation = useMutation({
		mutationFn: userAPI.requestBanReview,
		onSuccess: async () => {
			await queryClient.invalidateQueries({ queryKey: USER_STATE_KEY });
			toast.success(t("banReviewRequested"));
		},
		onError: () => toast.error(t("failedToRequestBanReview")),
	});

	return (
		<div className="flex h-full w-full items-center justify-center bg-card px-6 text-card-foreground">
			<div className="w-full max-w-md space-y-5 rounded-xl border border-destructive/30 bg-destructive/5 p-6 text-center shadow-sm">
				<div className="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-destructive/10">
					<ShieldAlert className="h-6 w-6 text-destructive" />
				</div>
				<div className="space-y-2">
					<h2 className="text-lg font-semibold">{t("banNoticeTitle")}</h2>
					<p className="text-sm text-muted-foreground">{t("banNoticeDescription")}</p>
				</div>
				<Button
					className="w-full"
					variant={reviewRequested ? "outline" : "default"}
					disabled={reviewRequested || reviewMutation.isPending}
					onClick={() => reviewMutation.mutate()}
				>
					{reviewRequested ? t("banReviewRequested") : t("banReviewSubmit")}
				</Button>
			</div>
		</div>
	);
}
