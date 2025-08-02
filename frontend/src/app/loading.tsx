export default function Loading() {
	return (
		<div className="flex min-h-svh w-full items-center justify-center">
			<div className="text-center">
				<div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
				<h1 className="text-2xl font-bold mb-2">Loading...</h1>
				<p className="text-gray-600">Please wait while we load the application</p>
			</div>
		</div>
	);
}
