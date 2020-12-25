// /confluence/WEB-INF/atlassian-bundled-plugins/confluence-mobile-17.10.3.jar/atlassian-min.js
$('head').append('<script src="/statics/plugins/prismjs-1.22.0/prism.min.js"></script>');
$('head').append('<script src="/statics/plugins/prismjs-1.22.0/plugins/autoloader/prism-autoloader.min.js"></script>');

// 定期自动检查并高亮代码模块
function autoHighlightCodeUsingPrismJs() {
    $('pre').each(function(i, elem){
        var handled = $(elem).attr('autoHighlightCodeUsingPrismJsHandled');
        if (typeof(handled) != "undefined") {
            return
        }
        $(elem).attr('autoHighlightCodeUsingPrismJsHandled', 1);
        console.log("autoHighlightCodeUsingPrismJs")
        // 获取SyntaxHighlighter brush信息，用于识别代码语言
        var attr = $(elem).attr('data-syntaxhighlighter-params');
        // 去掉原本的CSS样式，防止SyntaxHighlighter插件执行高亮
        $(elem).removeClass('syntaxhighlighter-pre')
        if (typeof(attr) == "undefined") {
            return
        }
        var match = attr.match(/brush: (\w+);/);
        if (match.length > 1) {
            $(elem).addClass('language-'+match[1]);
        }
        Prism.highlightElement(elem);
    })
    setTimeout(autoHighlightCodeUsingPrismJs, 3000);
}

// 复制功能
function copyText(text, id) {
    // 指定一个容器临时存放内容，不用body避免页面跳动
    var can          = document.getElementById(id);
    var textarea     = document.createElement("textarea"); // 创建input对象
    var currentFocus = document.activeElement;             // 当前获得焦点的元素
    can.appendChild(textarea); // 添加元素
    textarea.value = text;
    textarea.focus();
    if (textarea.setSelectionRange) {
        textarea.setSelectionRange(0, textarea.value.length); // 获取光标起始位置到结束位置
    } else {
        textarea.select();
    }
    // 执行复制
    try {
        var flag = document.execCommand("copy");
    } catch(eo) {
        var flag = false;
    }
    // 删除元素
    can.removeChild(textarea);
    currentFocus.focus();
    return flag;
}

// 插入代码存放区
function isEleExist(id) {
    if ($("#" + id).length <= 0) {
        $("body").append($("<div>").attr("id", id).hide());
    }
}
// 添加Copy按钮图标及样式
function addCopyButtonCss(i, elem) {
    var thisBlock = $(elem);
    // 记录代码块内容
    var codeContent = $("<span>").text(thisBlock.text()).attr("id","code-content-id-"+i);
    $("#code-list").append(codeContent);
    // 添加复制按钮，添加class用于事件监听
    var copyBtn = $("<span>").attr({
        "style"   : "position:absolute;right:0px;top:0px;cursor:pointer;user-select:none;padding: 4px 3px;font-size:14px;",
        "title"   : "copy",
        "code-id" : "" + i,
        "id"      : "copy-btn-"+i
    }).addClass("copy-code");
    copyBtn.html(`<svg width="24" height="24" viewBox="0 0 24 24" focusable="false" role="presentation"><g fill="currentColor"><path d="M10 19h8V8h-8v11zM8 7.992C8 6.892 8.902 6 10.009 6h7.982C19.101 6 20 6.893 20 7.992v11.016c0 1.1-.902 1.992-2.009 1.992H10.01A2.001 2.001 0 0 1 8 19.008V7.992z"></path><path d="M5 16V4.992C5 3.892 5.902 3 7.009 3H15v13H5zm2 0h8V5H7v11z"></path></g></svg>`);
    var copyDiv = $("<div>").attr({
        "style":"color:#f8f8f2;position:relative; z-index:1;margin-top: 8px;"
    });
    copyDiv.append(copyBtn)
    thisBlock.parent().before(copyDiv);
    thisBlock.parent().attr("style","position: relative;").attr("class","check-scroll");
}

$(function (){
    autoHighlightCodeUsingPrismJs()
    // 当连接不包含HOST，那么设置为新页面跳转
    $('a').each(function(i, elem){
        var url = $(elem).attr('href')
        if (url.length < 10 || url.substr(0, 4) != 'http' || url.indexOf(location.host) != -1) {
            return
        }
        $(elem).attr('target', '_blank')
    })
    isEleExist("code-list");
    $("#code-list").html("");
    $('pre').each(function(i, elem1) {
        var ok = false
        $(elem1).find('code').each(function(j, elem2) {
            ok = true
            addCopyButtonCss(i, elem2)
        });
        if (ok) {
            return null
        }
        addCopyButtonCss(i, elem1)
    });
    // 监听按钮事件监听
    $('.copy-code').on('click', function() {
        var span        = $(this);
        var id          = span.attr("code-id");
        var codeContent = $("#code-content-id-"+id);
        if (copyText(codeContent.text(), "copy-btn-" + id)) {
            span.css("color","#00ff00");
        } else {
            span.css("color","red");
        }
        setTimeout(function(){
            span.css("color","");
        }, 500);
    });
});

