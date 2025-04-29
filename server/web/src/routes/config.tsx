import { type RouteObject } from "react-router"
import { Navigate } from "react-router"
import { Suspense, lazy, ReactElement } from "react"
import { RoutePath, ROUTES } from "./constants"
import { useTranslation } from 'react-i18next'
import { TFunction } from 'i18next'
import { ProtectedRoute } from "@/feature/auth/components/ProtectedRoute"

// 直接导入布局组件
import { RootLayout } from "@/components/layout/root-layout"
import { MonitorPage } from "@/pages/monitor/layout"
import { RulesPage } from "@/pages/rule/layout"
import { SettingPage } from "@/pages/setting/layout"
import { LogAndEventPage } from "@/pages/logs/layout"

// 直接导入子组件
import GlobalSettingPage from "@/pages/setting/pages/global-setting/page"
import CertificatesPage from "@/pages/setting/pages/certificate/page"
import EventsPage from "@/pages/logs/pages/event/page"
import LogsPage from "@/pages/logs/pages/log/page"
import SiteManagerPage from "@/pages/setting/pages/site/page"
import { columns, MonitorOverview, payments } from "@/pages/monitor/components"
import { SysRules, UserRules, IpGroup } from "@/pages/rule/components"
import { LoadingFallback } from "@/components/common/loading-fallback"

// 懒加载认证页面
const LoginPage = lazy(() => import("@/pages/auth/login"))
const ResetPasswordPage = lazy(() => import("@/pages/auth/reset-password"))

// 懒加载组件包装器
const lazyLoad = (Component: React.ComponentType) => (
    <Suspense fallback={<LoadingFallback />}>
        <Component />
    </Suspense>
)

// 面包屑项类型定义
interface BreadcrumbItem {
    title: string
    path: string
    component: ReactElement
}

interface BreadcrumbConfig {
    defaultPath: string
    items: BreadcrumbItem[]
}

// 创建面包屑配置
export function createBreadcrumbConfig(t: TFunction): Record<RoutePath, BreadcrumbConfig> {
    return {
        [ROUTES.LOGS]: {
            defaultPath: "attack",
            items: [
                { title: t('breadcrumb.logs.attack'), path: "attack", component: <EventsPage /> },
                { title: t('breadcrumb.logs.protect'), path: "protect", component: <LogsPage /> },
            ]
        },
        [ROUTES.MONITOR]: {
            defaultPath: "overview",
            items: [
                { title: t('breadcrumb.monitor.overview'), path: "overview", component: <MonitorOverview columns={columns} data={payments} /> }
            ]
        },
        [ROUTES.RULES]: {
            defaultPath: "system",
            items: [
                { title: t('breadcrumb.rules.system'), path: "system", component: <SysRules /> },
                { title: t('breadcrumb.rules.user'), path: "user", component: <UserRules /> },
                { title: t('breadcrumb.rules.ipGroup'), path: "ip", component: <IpGroup /> }
            ]
        },
        [ROUTES.SETTINGS]: {
            defaultPath: "global",
            items: [
                { title: t('breadcrumb.settings.settings'), path: "global", component: <GlobalSettingPage /> },
                { title: t('breadcrumb.settings.siteManager'), path: "site", component: <SiteManagerPage /> },
                { title: t('breadcrumb.settings.certManager'), path: "cert", component: <CertificatesPage /> }
            ]
        }
    }
}

// 获取当前语言的面包屑配置
export function useBreadcrumbMap() {
    const { t } = useTranslation()
    return createBreadcrumbConfig(t)
}

// 生成子路由配置
function createChildRoutes(config: BreadcrumbConfig): RouteObject[] {
    return [
        {
            path: "",
            element: <Navigate to={config.defaultPath} replace />
        },
        ...config.items.map(item => ({
            path: item.path,
            element: item.component
        }))
    ]
}

// 路由配置
export function useRoutes(): RouteObject[] {
    const breadcrumbMap = useBreadcrumbMap()

    // 认证路由
    const authRoutes: RouteObject[] = [
        { path: "/login", element: lazyLoad(LoginPage) },
        { path: "/reset-password", element: lazyLoad(ResetPasswordPage) }
    ]

    // 应用路由
    const appRoutes: RouteObject = {
        element: <ProtectedRoute />,
        children: [{
            element: <RootLayout />,
            children: [
                {
                    path: "/",
                    element: <Navigate to={`${ROUTES.LOGS}/attack`} replace />
                },
                {
                    path: ROUTES.LOGS,
                    element: <LogAndEventPage />,
                    children: createChildRoutes(breadcrumbMap[ROUTES.LOGS])
                },
                {
                    path: ROUTES.MONITOR,
                    element: <MonitorPage />,
                    children: createChildRoutes(breadcrumbMap[ROUTES.MONITOR])
                },
                {
                    path: ROUTES.RULES,
                    element: <RulesPage />,
                    children: createChildRoutes(breadcrumbMap[ROUTES.RULES])
                },
                {
                    path: ROUTES.SETTINGS,
                    element: <SettingPage />,
                    children: createChildRoutes(breadcrumbMap[ROUTES.SETTINGS])
                }
            ]
        }]
    }

    return [...authRoutes, appRoutes]
}

// 默认面包屑配置，用于类型推断
export const breadcrumbMap = createBreadcrumbConfig(((key: string) => key) as unknown as TFunction) as ReturnType<typeof createBreadcrumbConfig>