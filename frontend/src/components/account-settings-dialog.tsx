"use client";

import * as React from "react";
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Checkbox } from "@/components/ui/checkbox";
import { toast } from "sonner";
import { useLanguage } from "@/contexts/theme";

interface AccountSettingsDialogProps {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	initialData: {
		age: number | null;
		gender: string;
	};
	onSave: (data: AccountSettingsDialogProps["initialData"]) => Promise<void> | void;
}

export function AccountSettingsDialog({
	open,
	onOpenChange,
	initialData,
	onSave,
}: AccountSettingsDialogProps) {
	const { t } = useLanguage();
	const [age, setAge] = React.useState(initialData.age);
	const [gender, setGender] = React.useState(() => {
		const validGenders = ["male", "female"];
		return validGenders.includes(initialData.gender) ? initialData.gender : "male";
	});
	const [isLoading, setIsLoading] = React.useState(false);

	React.useEffect(() => {
		if (open) {
			setAge(initialData.age);
			const validGenders = ["male", "female"];
			setGender(validGenders.includes(initialData.gender) ? initialData.gender : "male");
		}
	}, [open, initialData]);

	const handleSave = async () => {
		setIsLoading(true);

		try {
			await onSave({ age, gender });
			toast.success(t("userInfoSaveSuccess"));
			onOpenChange(false);
		} catch (error) {
			console.error("Error saving user settings:", error);
			toast.error(error instanceof Error ? error.message : t("pleaseTryAgain"));
		} finally {
			setIsLoading(false);
		}
	};

	return (
		<Dialog open={open} onOpenChange={onOpenChange}>
			<DialogContent className="z-[9998] h-fit max-h-[95vh] overflow-y-auto sm:max-w-[425px]">
				<DialogHeader>
					<DialogTitle>{t("accountSettings")}</DialogTitle>
					<DialogDescription>{t("accountSettingsDescription")}</DialogDescription>
				</DialogHeader>
				<div className="grid gap-4 py-4">
					<div className="grid grid-cols-4 items-center gap-4">
						<Label htmlFor="age" className="text-right">
							{t("age")}
						</Label>
						<Input
							id="age"
							type="number"
							value={age || ""}
							onChange={(e) =>
								setAge(e.target.value ? parseInt(e.target.value, 10) : null)
							}
							className="col-span-3"
							placeholder={t("agePlaceholder")}
							min="1"
							max="120"
						/>
					</div>
					<div className="grid grid-cols-4 items-center gap-4">
						<Label className="text-right">{t("gender")}</Label>
						<div className="col-span-3 flex gap-6">
							<div className="flex items-center space-x-2">
								<Checkbox
									id="male"
									checked={gender === "male"}
									onCheckedChange={(checked: boolean) => {
										if (checked) setGender("male");
									}}
								/>
								<Label htmlFor="male" className="flex items-center gap-1">
									{t("male")}
								</Label>
							</div>
							<div className="flex items-center space-x-2">
								<Checkbox
									id="female"
									checked={gender === "female"}
									onCheckedChange={(checked: boolean) => {
										if (checked) setGender("female");
									}}
								/>
								<Label htmlFor="female" className="flex items-center gap-1">
									{t("female")}
								</Label>
							</div>
						</div>
					</div>
				</div>
				<DialogFooter>
					<Button
						variant="outline"
						onClick={() => onOpenChange(false)}
						disabled={isLoading}
					>
						{t("cancel")}
					</Button>
					<Button onClick={handleSave} disabled={isLoading}>
						{isLoading ? t("saving") : t("saveChanges")}
					</Button>
				</DialogFooter>
			</DialogContent>
		</Dialog>
	);
}
