/*
 * Copyright (c) 2023 ivfzhou
 * csms is licensed under Mulan PSL v2.
 * You can use this software according to the terms and conditions of the Mulan PSL v2.
 * You may obtain a copy of Mulan PSL v2 at:
 *          http://license.coscl.org.cn/MulanPSL2
 * THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
 * EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
 * MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
 * See the Mulan PSL v2 for more details.
 */

/**
 * v-ripple 指令：为元素添加 Material Design 风格的波纹点击动效。
 *
 * 用法：在需要波纹效果的元素上添加 `v-ripple` 指令即可。
 *
 * 工作原理：
 * 1. 在元素挂载时，将元素的 position 设为 relative（若为 static），overflow 设为 hidden，确保波纹圆圈不会溢出元素边界。
 * 2. 监听 click 事件，根据鼠标点击位置计算波纹圆圈的圆心和直径（取元素宽高中较大者的 2 倍）。
 * 3. 创建一个绝对定位的圆形 span 元素，从圆心处通过 Web Animations API 播放缩放 + 淡出动画，（scale 0→1，opacity 0.5→0），持续时间 500ms，动画结束后移除该 span。
 * 4. 在元素卸载时，移除事件监听并恢复原始的 position 和 overflow 样式，清理内存引用。
 */
export default {
    mounted(el, binding, vnode) {
        const computedStyle = getComputedStyle(el)
        el._rippleOriginPosition = computedStyle.position
        el._rippleOriginOverflow = computedStyle.overflow
        if (computedStyle.position === 'static' || !computedStyle.position) el.style.position = 'relative'
        el.style.overflow = 'hidden'

        el._rippleHandler = (event) => {
            const rect = el.getBoundingClientRect()
            // 波纹圆圈的直径取元素宽高中较大者的 2 倍，确保动画覆盖整个元素表面。
            const size = Math.max(rect.width, rect.height) * 2
            // 以点击位置为圆心，计算圆圈左上角相对于元素的偏移坐标。
            const x = event.clientX - rect.left - size / 2
            const y = event.clientY - rect.top - size / 2

            const span = document.createElement('span')
            span.style.cssText = [
                `position: absolute`,
                `left: ${x}px`,
                `top: ${y}px`,
                `width: ${size}px`,
                `height: ${size}px`,
                `border-radius: 50%`,
                `background-color: rgba(0, 0, 0, 0.12)`,
                `pointer-events: none`
            ].join(';')
            el.appendChild(span)
            const animation = span.animate(
                [
                    {transform: 'scale(0)', opacity: 0.5},
                    {transform: 'scale(1)', opacity: 0}
                ],
                {
                    duration: 500,
                    easing: 'ease-out',
                    fill: 'forwards'
                }
            )
            animation.onfinish = () => span.remove()
        }
        el.addEventListener('click', el._rippleHandler)
    },
    unmounted(el, binding, vnode) {
        el.removeEventListener('click', el._rippleHandler)
        el.style.overflow = el._rippleOriginOverflow
        el.style.position = el._rippleOriginPosition
        delete el._rippleOriginPosition
        delete el._rippleOriginOverflow
        delete el._rippleHandler
    }
}
