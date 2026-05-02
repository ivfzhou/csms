import clipboard
import json
import os
import time
from pywinauto import base_wrapper
from pywinauto import findwindows
from pywinauto.controls import hwndwrapper
from pywinauto.controls import win32_controls
from uiautomation import DocumentControl


class Rsp:
    result = False
    message = ''
    content = ''

    def to_json(self):
        return json.dumps(self, default=lambda o: o.__dict__)


def handle():
    while True:
        time.sleep(10)
        try:
            with open(r'C:\dialog_conversation.txt', 'r+', encoding='utf-8') as file:
                content = file.read()
                lines = content.split('\n')
                line = lines[len(lines) - 1]
                if len(lines) % 2 == 1 and content != '' and line.startswith('ask: '):
                    rsp = ''
                    data = json.loads(line[4:])
                    method = data['method']
                    if method == 'get_dialog':
                        rsp = get_dialog(data['title']).to_json()
                    elif method == 'click_button':
                        rsp = click_button(data['title'], data['button'], data['index']).to_json()
                    elif method == 'list_window':
                        rsp = list_window().to_json()
                    elif method == 'close_windows':
                        rsp = close_windows(data['title'], data['index']).to_json()
                    elif method == 'get_cmd_content':
                        rsp = get_cmd_content(data['index']).to_json()
                    elif method == 'send_cmd_content':
                        rsp = send_cmd_content(data['content'], data['index']).to_json()
                    else:
                        print('no title match ' + method)
                    if rsp != '':
                        file.seek(0, 2)
                        file.write('\nanswer: ')
                        file.write(rsp)
                file.close()
        except BaseException as e:  # pylint: disable=broad-except
            print(e)


def get_dialog(title):
    print('get dialog ' + title)
    res = Rsp()
    content = {}
    try:
        cur = 0
        res.result = True
        res.content = content
        dialogs = findwindows.find_windows()
        for v in dialogs:
            hw = hwndwrapper.HwndWrapper(v)
            if hw.window_text() == title:
                children = hw.children()
                txt = ''
                for child in children:
                    text = str(base_wrapper.BaseWrapper.window_text(child))
                    txt += text
                content[cur] = txt
                cur += 1
    except BaseException as e:  # pylint: disable=broad-except
        res = Rsp()
        res.message = repr(e)
    return res


def click_button(title, but, index):
    print("click button " + title + " " + but)
    res = Rsp()
    try:
        cur = 0
        dialogs = findwindows.find_windows()
        for v in dialogs:
            hw = hwndwrapper.HwndWrapper(v)
            if hw.window_text() == title:
                if cur != index:
                    cur += 1
                    continue
                children = hw.children()
                for child in children:
                    text = str(base_wrapper.BaseWrapper.window_text(child))
                    if text.__contains__(but):
                        win32_controls.ButtonWrapper.click(child, double=True)
                        res.result = True
                        return res
                return res
    except BaseException as e:  # pylint: disable=broad-except
        res = Rsp()
        res.message = repr(e)
    return res


def list_window():
    print('list window')
    res = Rsp()
    try:
        res.result = True
        res.content = []
        dialogs = findwindows.find_windows()
        for v in dialogs:
            hw = hwndwrapper.HwndWrapper(v)
            res.content.append(hw.window_text())
    except BaseException as e:  # pylint: disable=broad-except
        res = Rsp()
        res.message = repr(e)
    return res


def close_windows(title, index):
    print('close window ' + title)
    res = Rsp()
    try:
        cur = 0
        dialogs = findwindows.find_windows()
        for v in dialogs:
            hw = hwndwrapper.HwndWrapper(v)
            if hw.window_text() == title:
                if cur != index:
                    cur += 1
                    continue
                hw.close()
                res.result = True
                return res
    except BaseException as e:  # pylint: disable=broad-except
        res = Rsp()
        res.message = repr(e)
    return res


def get_cmd_content(index):
    print('get cmd content')
    res = Rsp()
    try:
        res.result = True
        window = DocumentControl(Name="Text Area", searchDepth=3, foundIndex=index)
        if window.Exists():
            window.SendKeys('{Ctrl}A')
            window.SendKeys('{Ctrl}C')
            data = clipboard.paste()
            res.content = str(data)
            res.content = res.content.replace('\n', r'\n')
    except BaseException as e:  # pylint: disable=broad-except
        res = Rsp()
        res.message = repr(e)
    return res


def send_cmd_content(content, index):
    print('send cmd content ' + content)
    res = Rsp()
    try:
        res.result = True
        window = DocumentControl(Name="Text Area", searchDepth=3, foundIndex=index)
        window.SendKeys(content)
    except BaseException as e:  # pylint: disable=broad-except
        res = Rsp()
        res.message = repr(e)
    return res


def main():
    username = os.getlogin()
    print('run handler in ' + username)
    handle()


main()
