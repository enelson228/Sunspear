import { ref, onUnmounted } from 'vue'
import { useAuthStore } from '@/stores/auth'

export function useWebSocket(path) {
    const data = ref(null)
    const connected = ref(false)
    const error = ref(null)

    let ws = null
    let reconnectTimeout = null
    let reconnectDelay = 1000
    const maxReconnectDelay = 30000

    function connect() {
        const authStore = useAuthStore()
        const token = authStore.getToken()

        if (!token) {
            error.value = 'No authentication token available'
            return
        }

        // Construct WebSocket URL from API base URL
        const apiBaseUrl = import.meta.env.VITE_API_URL || 'http://localhost:8080/api'
        const wsUrl = apiBaseUrl
            .replace(/^http:/, 'ws:')
            .replace(/^https:/, 'wss:')
            .replace(/\/api$/, '') // Remove /api suffix if present

        const fullUrl = `${wsUrl}${path}?token=${encodeURIComponent(token)}`

        try {
            ws = new WebSocket(fullUrl)

            ws.onopen = () => {
                connected.value = true
                error.value = null
                reconnectDelay = 1000 // Reset reconnect delay on successful connection
            }

            ws.onmessage = (event) => {
                try {
                    data.value = JSON.parse(event.data)
                } catch (err) {
                    console.error('Failed to parse WebSocket message:', err)
                    error.value = 'Invalid message format'
                }
            }

            ws.onerror = (err) => {
                error.value = 'WebSocket connection error'
                console.error('WebSocket error:', err)
            }

            ws.onclose = () => {
                connected.value = false

                // Auto-reconnect with exponential backoff
                reconnectTimeout = setTimeout(() => {
                    reconnectDelay = Math.min(reconnectDelay * 2, maxReconnectDelay)
                    connect()
                }, reconnectDelay)
            }
        } catch (err) {
            error.value = err.message
            console.error('WebSocket connection failed:', err)
        }
    }

    function disconnect() {
        if (reconnectTimeout) {
            clearTimeout(reconnectTimeout)
            reconnectTimeout = null
        }

        if (ws) {
            ws.close()
            ws = null
        }

        connected.value = false
    }

    // Clean up on component unmount
    onUnmounted(() => {
        disconnect()
    })

    return {
        data,
        connected,
        error,
        connect,
        disconnect
    }
}
