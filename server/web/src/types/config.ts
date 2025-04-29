export interface AppConfig {
    name: string
    directives: string
    logLevel?: string
    logFormat?: string
    logFile?: string
    transactionTTL?: number
}

export interface EngineConfig {
    bind: string
    useBuiltinRules: boolean
    appConfig: AppConfig[]
}

export interface HaproxyConfig {
    thread: number
    configBaseDir: string
    haproxyBin: string
    backupsNumber?: number
    spoeAgentAddr?: string
    spoeAgentPort?: number
}

export interface ConfigResponse {
    id: string
    name: string
    isDebug: boolean
    isResponseCheck: boolean
    engine: EngineConfig
    haproxy: HaproxyConfig
    createdAt: string
    updatedAt: string
}

export interface ConfigPatchRequest {
    name?: string
    isDebug?: boolean
    isResponseCheck?: boolean
    engine?: {
        bind?: string
        useBuiltinRules?: boolean
        appConfig?: {
            name?: string
            directives?: string
            logLevel?: string
            logFormat?: string
            logFile?: string
            transactionTTL?: number
        }[]
    }
    haproxy?: {
        thread?: number
        configBaseDir?: string
        haproxyBin?: string
        backupsNumber?: number
        spoeAgentAddr?: string
        spoeAgentPort?: number
    }
}

