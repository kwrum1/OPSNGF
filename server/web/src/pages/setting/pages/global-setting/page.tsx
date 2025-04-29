// src/pages/setting/pages/global-setting/page.tsx
import { useEffect } from 'react'
import { useConfigQuery } from '@/feature/global-setting/hooks/useConfig'
import { useRunnerStatusQuery, useRunnerControl } from '@/feature/global-setting/hooks/useRunner'
import { EngineStatus } from '@/feature/global-setting/components/EngineStatus'
import { ConfigForm } from '@/feature/global-setting/components/ConfigForm'
import { Settings } from 'lucide-react'
import { AdvancedErrorDisplay } from '@/components/common/error/errorDisplay'
import { Card } from '@/components/ui/card'
import { AnimatedContainer } from '@/components/ui/animation/components/animated-container'
import { useTranslation } from 'react-i18next'

export default function GlobalSettingPage() {
    const { t } = useTranslation()
    
    // 获取配置数据
    const {
        config,
        isLoading: isConfigLoading,
        error: configError,
        refetch: refetchConfig
    } = useConfigQuery()

    // 获取运行器状态
    const {
        status,
        isLoading: isStatusLoading,
        error: statusError,
        refetch: refetchStatus
    } = useRunnerStatusQuery()

    // 运行器控制
    const { controlRunner, isLoading: isControlLoading, error: controlError, clearError: clearControlError } = useRunnerControl()

    // 当页面加载时获取最新配置
    useEffect(() => {
        // 页面加载时，获取最新配置和状态
        refetchConfig()
        refetchStatus()
    }, [refetchConfig, refetchStatus])

    // 运行器控制处理函数
    const handleStart = () => {
        clearControlError()
        controlRunner('start')
    }
    const handleStop = () => {
        clearControlError()
        controlRunner('stop')
    }
    const handleRestart = () => {
        clearControlError()
        controlRunner('restart')
    }
    const handleForceStop = () => {
        clearControlError()
        controlRunner('force_stop')
    }
    const handleReload = () => {
        clearControlError()
        controlRunner('reload')
    }

    // 根据错误类型选择适当的重试函数
    const handleRetry = () => {
        if (configError) refetchConfig()
        if (statusError) refetchStatus()
    }

    return (
        <AnimatedContainer variant="smooth" className="h-full overflow-y-auto hide-scrollbar p-0">
            <div className="container p-6  max-w-2xl flex flex-col gap-10">
                {/* header */}
                <Card className="border-none shadow-none gap-4 flex flex-col bg-zinc-50 rounded-md p-4">
                    <div className="flex items-center gap-2">
                        <Settings className="h-5 w-5 text-primary" />
                        <h2 className="text-xl font-semibold">{t("globalSetting.title")}</h2>
                    </div>
                    <p className="text-muted-foreground">
                        {t("globalSetting.description")}
                    </p>
                </Card>


                {/* 错误处理：优先显示配置错误，其次显示状态错误 */}
                {(configError || statusError) && (
                    <AdvancedErrorDisplay
                        error={configError || statusError}
                        onRetry={handleRetry}
                    />
                )}
                {controlError && (
                    <AdvancedErrorDisplay
                        error={controlError}
                    />
                )}

                {/* 引擎状态和配置表单 */}

                <div className="pb-6  border-none shadow-none flex flex-col gap-8">
                    <EngineStatus
                        status={status}
                        isLoading={isStatusLoading}
                        onStart={handleStart}
                        onStop={handleStop}
                        onRestart={handleRestart}
                        onForceStop={handleForceStop}
                        onReload={handleReload}
                        isControlLoading={isControlLoading}
                    />

                    <ConfigForm
                        config={config}
                        isLoading={isConfigLoading}
                    />
                </div>
            </div>
        </AnimatedContainer>
    )
}