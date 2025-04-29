// Loading component with translations
export const LoadingFallback = () => {
    return (
        <div className="fixed inset-0 flex items-center justify-center z-[9999]">
            <div className="absolute inset-0 bg-[size:200%_200%] animate-aurora">
                <div className="absolute inset-0 overflow-hidden">
                    <div className="absolute w-[80%] h-[80%] top-[10%] left-[10%] bg-purple-300/20 rounded-full blur-3xl animate-float"></div>
                    <div className="absolute w-[40%] h-[40%] top-[5%] right-[15%] bg-cyan-300/20 rounded-full blur-3xl animate-float-reverse"></div>
                    <div className="absolute w-[50%] h-[50%] bottom-[5%] left-[15%] bg-indigo-300/20 rounded-full blur-3xl animate-pulse-glow"></div>
                </div>
            </div>
            <img
                src="/logo.svg"
                alt="Loading RuiQi WAF"
                className="relative w-20 h-20 animate-logo-pulse z-10"
            />
        </div>
    )
}