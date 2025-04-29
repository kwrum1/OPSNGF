import { Link, useLocation, useNavigate } from "react-router"
import { cn } from "@/lib/utils"
import { Settings, Shield, BarChart2, FileText, LogOut } from "lucide-react"
import { ROUTES } from "@/routes/constants"
import { Card, CardHeader, CardContent, CardFooter, CardTitle, CardDescription } from "@/components/ui/card"
import { useTranslation } from 'react-i18next'
import { TFunction } from 'i18next'
import { useAuthStore } from "@/store/auth"

// 为每个导航项添加 display 选项
function createSidebarConfig(t: TFunction) {
    return [
        {
            title: t('sidebar.monitor'),
            icon: BarChart2,
            href: ROUTES.MONITOR,
            display: true  // 控制是否显示此导航项
        },
        {
            title: t('sidebar.logs'),
            icon: FileText,
            href: ROUTES.LOGS,
            display: true  // 控制是否显示此导航项
        },
        {
            title: t('sidebar.rules'),
            icon: Shield,
            href: ROUTES.RULES,
            display: true  // 控制是否显示此导航项
        },
        {
            title: t('sidebar.settings'),
            icon: Settings,
            href: ROUTES.SETTINGS,
            display: true  // 控制是否显示此导航项
        }
    ] as const
}

// 添加一个配置接口，允许传入对象来控制各导航项显示状态
interface SidebarDisplayConfig {
    monitor?: boolean
    logs?: boolean
    rules?: boolean
    settings?: boolean
}

interface SidebarProps {
    displayConfig?: SidebarDisplayConfig
}

export function Sidebar({ displayConfig = {} }: SidebarProps) {
    const location = useLocation()
    const { t } = useTranslation()
    const navigate = useNavigate()
    const { logout } = useAuthStore()

    // 获取当前路径的第一级
    const currentFirstLevelPath = '/' + location.pathname.split('/')[1]

    // 使用 t 函数生成 sidebarItems，并应用 display 配置
    const sidebarItems = createSidebarConfig(t).map(item => {
        // 根据路径名确定哪个配置属性
        let configKey: keyof SidebarDisplayConfig = 'monitor'
        if (item.href === ROUTES.LOGS) configKey = 'logs'
        if (item.href === ROUTES.RULES) configKey = 'rules'
        if (item.href === ROUTES.SETTINGS) configKey = 'settings'

        // 如果配置中指定了该项的显示状态，则使用配置的值
        // 否则使用默认值 true
        const shouldDisplay = displayConfig[configKey] !== undefined ?
            displayConfig[configKey] : item.display

        return {
            ...item,
            display: shouldDisplay
        }
    })

    const handleLogout = () => {
        logout()
        navigate('/login')
    }

    return (
        <Card className="w-[17.69rem] min-w-[17.69rem] flex flex-col rounded-none gap-1 border-0 shadow-none overflow-auto">
            <CardHeader className="pt-[0.0625rem] pb-0 gap-5 w-full items-center justify-center space-y-0 ">
                <CardTitle
                    className="w-[5rem] h-[5rem] rounded-full flex justify-center items-center"
                >
                    <img src="/logo.svg" alt="logo" className="w-[5rem] h-[5rem] rounded-full" />
                </CardTitle>
                {/* <CardDescription className="text-[1.75rem] font-bold leading-[1.4] tracking-[0.0125rem] normal-case bg-clip-text text-transparent bg-gradient-to-r from-purple-700 to-cyan-300 drop-shadow-sm"> */}
                <CardDescription className="text-[1.75rem] font-bold leading-[1.4] tracking-[0.0125rem] normal-case bg-clip-text text-transparent bg-gradient-to-r from-purple-600 via-cyan-400 to-purple-300 drop-shadow-sm">
                    {t('sidebar.title')}
                </CardDescription>
            </CardHeader>

            <CardContent className="pt-[6rem] pl-[3rem] pb-0 pr-0">
                <nav className="flex flex-col gap-[1.125rem]">
                    {sidebarItems
                        .filter(item => item.display) // 只显示 display 为 true 的项
                        .map((item) => {
                            const isActive = currentFirstLevelPath === item.href
                            return (
                                <Link
                                    key={item.href}
                                    to={item.href}
                                    className="flex items-center gap-[1.125rem] group"
                                >
                                    <div className={cn(
                                        "p-2 rounded-md w-[3.5rem] h-[3.5rem]",
                                        "transform transition-all duration-500 ease-out",
                                        isActive
                                            ? "bg-zinc-600 scale-110"
                                            : "bg-gray-100 group-hover:scale-105 group-hover:bg-gray-900/20"
                                    )}>
                                        <item.icon
                                            strokeWidth={1}
                                            className={cn(
                                                "w-full h-full shrink-0",
                                                "transform transition-all duration-500 ease-out",
                                                isActive ? "stroke-white animate-icon-shake" : "stroke-gray-500 group-hover:stroke-gray-900"
                                            )}
                                        />
                                    </div>
                                    <span className={cn(
                                        "text-[1.5rem] leading-[1.6] tracking-[0.0625rem] text-gray-600",
                                        "transform transition-all duration-500 ease-out",
                                        isActive
                                            ? "font-bold translate-x-2"
                                            : "font-normal group-hover:translate-x-1"
                                    )}>
                                        {item.title}
                                    </span>
                                </Link>
                            )
                        })}
                </nav>
            </CardContent>

            <CardFooter className="pt-[6.25rem] pl-[3rem] pb-0 pr-0">
                <div className="flex items-center gap-[1.125rem] cursor-pointer group" onClick={handleLogout}>
                    <div className={cn(
                        "p-2 rounded-md w-[3.5rem] h-[3.5rem]",
                        "transform transition-all duration-500 ease-out",
                        "bg-gray-100 group-hover:bg-gray-900/20 group-hover:scale-105",
                        "group-active:bg-gray-900"
                    )}>
                        <LogOut strokeWidth={1}
                            className={cn(
                                "w-full h-full shrink-0",
                                "transform transition-all duration-500 ease-out",
                                "stroke-gray-500 group-hover:stroke-gray-900",
                                "group-active:stroke-white"
                            )} />
                    </div>
                    <span className={cn(
                        "text-[1.5rem] leading-[1.6] tracking-[0.0625rem] text-gray-600",
                        "transform transition-all duration-500 ease-out",
                        "font-normal group-hover:translate-x-1",
                        "group-active:font-bold group-active:translate-x-2"
                    )}>
                        {t('sidebar.logout')}
                    </span>
                </div>
            </CardFooter>
        </Card>
    )
} 