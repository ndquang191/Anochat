import { ImageResponse } from "next/og";

export const alt = "AnoChat — Nói điều thật lòng";
export const size = { width: 1200, height: 630 };
export const contentType = "image/png";

function GhostLogo() {
	return (
		<svg width="156" height="156" viewBox="0 0 32 32" aria-hidden="true">
			<path
				d="M3 13 A13 13 0 0 1 29 13 L29 27 Q24.3 33 19.7 27 Q16 33 12.3 27 Q7.7 33 3 27 Z"
				fill="#516b91"
			/>
			<circle cx="11.5" cy="16" r="2.4" fill="white" />
			<circle cx="20.5" cy="16" r="2.4" fill="white" />
			<circle cx="12.2" cy="16.8" r="1.2" fill="#516b91" />
			<circle cx="21.2" cy="16.8" r="1.2" fill="#516b91" />
			<path
				d="M13 21.5 Q16 24 19 21.5"
				stroke="white"
				strokeWidth="1.4"
				strokeLinecap="round"
				fill="none"
			/>
		</svg>
	);
}

export default function OpenGraphImage() {
	return new ImageResponse(
		<div
			style={{
				display: "flex",
				position: "relative",
				width: "100%",
				height: "100%",
				overflow: "hidden",
				background: "#f7fafc",
				color: "#263a56",
				fontFamily: "Arial, sans-serif",
			}}
		>
			<div
				style={{
					position: "absolute",
					width: 520,
					height: 520,
					left: -180,
					top: -245,
					borderRadius: "50%",
					background: "#dce8f6",
				}}
			/>
			<div
				style={{
					position: "absolute",
					width: 440,
					height: 440,
					right: -105,
					bottom: -245,
					borderRadius: "50%",
					background: "#eee3f6",
				}}
			/>

			<div
				style={{
					display: "flex",
					alignItems: "center",
					width: "100%",
					padding: "76px 88px",
				}}
			>
				<div
					style={{
						display: "flex",
						alignItems: "center",
						justifyContent: "center",
						width: 230,
						height: 230,
						flexShrink: 0,
						borderRadius: 58,
						background: "white",
						boxShadow: "0 24px 70px rgba(81, 107, 145, 0.18)",
						border: "2px solid rgba(81, 107, 145, 0.10)",
					}}
				>
					<GhostLogo />
				</div>

				<div style={{ display: "flex", flexDirection: "column", marginLeft: 70 }}>
					<div
						style={{
							display: "flex",
							fontSize: 66,
							fontWeight: 800,
							letterSpacing: "-2px",
							color: "#516b91",
						}}
					>
						AnoChat
					</div>
					<div
						style={{
							display: "flex",
							maxWidth: 690,
							marginTop: 20,
							fontSize: 44,
							fontWeight: 700,
							lineHeight: 1.18,
							letterSpacing: "-1px",
						}}
					>
						Gặp một người lạ. Nói điều thật lòng.
					</div>
					<div
						style={{
							display: "flex",
							marginTop: 27,
							fontSize: 24,
							color: "#64748b",
						}}
					>
						Riêng tư · Không profile · Không áp lực
					</div>
				</div>
			</div>
		</div>,
		size,
	);
}
