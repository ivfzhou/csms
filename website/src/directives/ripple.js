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

export default {
    mounted(el, binding, vnode) {
        const computedStyle = getComputedStyle(el)
        el._rippleOriginPosition = computedStyle.position
        el._rippleOriginOverflow = computedStyle.overflow
        if (computedStyle.position === 'static' || !computedStyle.position) el.style.position = 'relative'
        el.style.overflow = 'hidden'

        el._rippleHandler = (event) => {
            const rect = el.getBoundingClientRect()
            const size = Math.max(rect.width, rect.height) * 2
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
