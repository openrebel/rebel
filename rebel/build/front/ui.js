const isSecure = window.location.href.toLowerCase().startsWith("https://");
const onMobile = (/Android|webOS|iPhone|iPad|iPod|IEMobile|Opera Mini/i.test(navigator.userAgent));

const favicon   = document.getElementById("favicon");

const container = document.getElementById("container");
const iconsbar  = document.getElementById("iconsbar");
const menu      = document.getElementById("menu");
const btnMenu   = document.getElementById("floatingbutton");
const divIcon   = document.getElementById("logo");
const cap       = document.getElementById("cap");

const searchbox = document.getElementById("searchbox");
const txtSearch = document.getElementById("txtSearch");
const btnSearchClear = document.getElementById("btnSearchClear");

const btnSettings = document.getElementById("btnSettings");
const btnLogoff   = document.getElementById("btnLogout");

let menu_isopen = false;
let menu_button_drag = false;
let menu_button_moved = false;
let menu_startPos = [0, 0];
let menu_lastAltPress = 0;

(function() {
    let menu_button_pos = JSON.parse(localStorage.getItem("menu_button_pos"));

    if (menu_button_pos) {
        btnMenu.style.borderRadius = menu_button_pos.borderRadius;
        btnMenu.style.left = menu_button_pos.left;
        btnMenu.style.top = menu_button_pos.top;
        btnMenu.style.width = menu_button_pos.width;
        btnMenu.style.height = menu_button_pos.height;
    
        divIcon.style.left = menu_button_pos.l_left;
        divIcon.style.top = menu_button_pos.l_top;
        divIcon.style.width = menu_button_pos.l_width;
        divIcon.style.height = menu_button_pos.l_height;
    }

    Menu_UpdatePosition();
})();

function RgbToHsl(color) {
    let r = color[0] / 255;
    let g = color[1] / 255;
    let b = color[2] / 255;

    let cmin = Math.min(r, g, b);
    let cmax = Math.max(r, g, b);
    let delta = cmax - cmin;

    let h, s, l;

    if (delta == 0) h = 0;
    else if (cmax == r) h = ((g - b) / delta) % 6;
    else if (cmax == g) h = (b - r) / delta + 2;
    else h = (r - g) / delta + 4;

    h = Math.round(h * 60);

    if (h < 0) h += 360;

    l = (cmax + cmin) / 2;
    s = delta == 0 ? 0 : delta / (1 - Math.abs(2 * l - 1));
    s = +(s * 100).toFixed(1);
    l = +(l * 100).toFixed(1);

    return [h, s, l];
}
function SetAccentColor(accent) {
    let rgbString = `rgb(${accent[0]},${accent[1]},${accent[2]})`;
    let hsl = this.RgbToHsl(accent);

    let step1 = `hsl(${hsl[0]-4},${hsl[1]}%,${hsl[2]*.78}%)`;
    let step2 = `hsl(${hsl[0]+7},${hsl[1]}%,${hsl[2]*.9}%)`; //--select-color
    let step3 = `hsl(${hsl[0]-4},${hsl[1]}%,${hsl[2]*.8}%)`;

    let root = document.documentElement;
    root.style.setProperty("--accent-a", step2);
    root.style.setProperty("--accent-b", step3);
    root.style.setProperty("--accent-c", rgbString);

    let ico = "<svg version=\"1.1\" xmlns:serif=\"http://www.serif.com/\" xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" x=\"0px\" y=\"0px\" width=\"48px\" height=\"48px\"  viewBox=\"0 0 48 48\" enable-background=\"new 0 0 48 48\" xml:space=\"preserve\">"+
        "<g fill=\""+step2+"\" transform=\"matrix(-1,0,0,1,48,0)\">"+
        "<path d=\"M26.935,0.837h7.491l0.624,14.984l-8.24,1.873L26.935,0.837z\"/>"+
        "<path d=\"M38.172,19.068l-3.871,8.866l-22.974,9.489l0.125-8.44l13.412-2.299V15.821L1.712,20.566l1.998,26.221 l42.579,0.375l-0.249-30.466L38.172,19.068z\"/>"+
        "<path d=\"M4.459,0.837l0.374,16.857l8.741-1.873l-0.5-14.984H4.459z\"/>"+
        "<path d=\"M15.821,0.837h7.304L24,13.2l-8.054,1.498L15.821,0.837z\"/>"+
        "<path d=\"M37.672,0.837h7.367l1.249,12.986l-8.491,1.998L37.672,0.837z\"/>"+
        "</g></svg>";

    favicon.href = "data:image/svg+xml;base64," + btoa(ico);
}

function Menu_UpdatePosition() {
    menu.style.visibility = menu_isopen ? "visible" : "hidden";
    cap.style.visibility = menu_isopen ? "visible" : "hidden";

    let left = parseInt(btnMenu.style.left);

    if (btnMenu.style.left == "0px" || left < 10 || btnMenu.style.top == "") {
        menu.style.left = "20px";
        menu.style.top = "20px";
        menu.style.bottom = "20px";
        menu.style.transform = menu_isopen ? "none" : "translateX(calc(-100% - 24px))";

    } else if (btnMenu.style.left == "calc(100% - 48px)" || left > 90) {
        menu.style.left = "calc(100% - var(--menu-width) - 20px)";
        menu.style.top = "20px";
        menu.style.bottom = "20px";
        menu.style.transform = menu_isopen ? "none" : "translateX(100%)";

    } else {        
        menu.style.left = `max(20px, min(calc(${left}% - var(--menu-width) / 2) + 32px, calc(100% - var(--menu-width) - 20px)))`;
                
        if (btnMenu.style.top == "0px") {
            menu.style.top = "20px";
            menu.style.bottom = "min(200px, 20%)";
            menu.style.transform = menu_isopen ? "none" : "translateY(-100%)";
        } else {
            menu.style.top = "min(200px, 20%)";
            menu.style.bottom = "20px";
            menu.style.transform = menu_isopen ? "none" : "translateY(+100%)";
        }    
    }
};

function Menu_Open() {
    menu_isopen = true;
    Menu_UpdatePosition();

    if (menu_isopen) {
        setTimeout(()=>{ txtSearch.focus(); }, 150);
    }
};
function Menu_Close() {
    menu_isopen = false;
    Menu_UpdatePosition();
};
function Menu_Toogle() {
    menu_isopen = !menu_isopen;
    Menu_UpdatePosition();

    if (menu_isopen) {
        setTimeout(()=>{ txtSearch.focus(); }, 150);
    }
};

document.body.onkeyup = event => {
    if (event.code == "AltLeft") {
        event.preventDefault();
        if (Date.now() - menu_lastAltPress < 250) {
            menu_lastAltPress = 0;
            Menu_Toogle();
        } else {
            menu_lastAltPress = Date.now();
        }
    } else
    menu_lastAltPress = 0;
};

document.body.addEventListener("mousemove", event => {
    if (event.buttons != 1) menu_button_drag = false;
    if (!menu_button_drag) return;

    if (Math.abs(menu_startPos[0] - event.clientX) > 2 || Math.abs(menu_startPos[1] - event.clientY) > 2) {
        menu_button_moved = true;
    }

    let px = event.x / container.clientWidth;
    let py = event.y / container.clientHeight;

    if (event.x < 56 && event.y < 56) {
        btnMenu.style.borderRadius = "4px 8px 48px 8px";
        btnMenu.style.left = "0px";
        btnMenu.style.top = "0px";
        btnMenu.style.width = "48px";
        btnMenu.style.height = "48px";

        divIcon.style.left = "8px";
        divIcon.style.top = "6px";
        divIcon.style.width = "26px";
        divIcon.style.height = "26px";

    } else if (event.x < 56 && event.y > container.clientHeight - 48) {
        btnMenu.style.borderRadius = "8px 48px 8px 4px";
        btnMenu.style.left = "0px";
        btnMenu.style.top = "calc(100% - 48px)";
        btnMenu.style.width = "48px";
        btnMenu.style.height = "48px";

        divIcon.style.left = "8px";
        divIcon.style.top = "16px";
        divIcon.style.width = "26px";
        divIcon.style.height = "26px";

    } else if (event.x > container.clientWidth - 48 && event.y < 56) {
        btnMenu.style.borderRadius = "8px 4px 8px 64px";
        btnMenu.style.left = "calc(100% - 48px)";
        btnMenu.style.top = "0px";
        btnMenu.style.width = "48px";
        btnMenu.style.height = "48px";

        divIcon.style.left = "16px";
        divIcon.style.top = "6px";
        divIcon.style.width = "26px";
        divIcon.style.height = "26px";

    } else if (event.x > container.clientWidth - 48 && event.y > container.clientHeight - 48) {
        btnMenu.style.borderRadius = "64px 8px 4px 8px";
        btnMenu.style.left = "calc(100% - 48px)";
        btnMenu.style.top = "calc(100% - 48px)";
        btnMenu.style.width = "48px";
        btnMenu.style.height = "48px";

        divIcon.style.left = "16px";
        divIcon.style.top = "16px";
        divIcon.style.width = "26px";
        divIcon.style.height = "26px";

    } else if (px < py && 1-px > py) { //left
        let y = 100 * (event.y - 32) / container.clientHeight;

        btnMenu.style.borderRadius = "14px 40px 40px 14px";
        btnMenu.style.left = "0px";
        btnMenu.style.top = `${y}%`;
        btnMenu.style.width = "48px";
        btnMenu.style.height = "64px";

        divIcon.style.left = "8px";
        divIcon.style.top = "18px";
        divIcon.style.width = "28px";
        divIcon.style.height = "28px";

    } else if (px > py && 1-px > py) { //top
        let x = 100 * (event.x - 32) / container.clientWidth;

        btnMenu.style.borderRadius = "14px 14px 40px 40px";
        btnMenu.style.left = `${x}%`;
        btnMenu.style.top = "0px";
        btnMenu.style.width = "64px";
        btnMenu.style.height = "48px";

        divIcon.style.left = "19px";
        divIcon.style.top = "6px";
        divIcon.style.width = "28px";
        divIcon.style.height = "28px";

    } else if (px < py && 1-px < py) { //bottom
        let x = 100 * (event.x - 32) / container.clientWidth;

        btnMenu.style.borderRadius = "40px 40px 14px 14px";
        btnMenu.style.left = `${x}%`;
        btnMenu.style.top = "calc(100% - 48px)";
        btnMenu.style.width = "64px";
        btnMenu.style.height = "48px";

        divIcon.style.left = "19px";
        divIcon.style.top = "16px";
        divIcon.style.width = "28px";
        divIcon.style.height = "28px";

    } else if (px > py && 1-px < py) { //right
        let y = 100 * (event.y - 32) / container.clientHeight;

        btnMenu.style.borderRadius = "40px 14px 14px 40px";
        btnMenu.style.left = "calc(100% - 48px)";
        btnMenu.style.top = `${y}%`;
        btnMenu.style.width = "48px";
        btnMenu.style.height = "64px";

        divIcon.style.left = "14px";
        divIcon.style.top = "18px";
        divIcon.style.width = "28px";
        divIcon.style.height = "28px";
    }

    Menu_UpdatePosition();
});

document.body.addEventListener("mouseup", event => {
    if (menu_button_moved) {
        localStorage.setItem("menu_button_pos", JSON.stringify({
            borderRadius: btnMenu.style.borderRadius,
            left:         btnMenu.style.left,
            top:          btnMenu.style.top,
            width:        btnMenu.style.width,
            height:       btnMenu.style.height,
            l_left:       divIcon.style.left,
            l_top:        divIcon.style.top,
            l_width:      divIcon.style.width,
            l_height:     divIcon.style.height
        }));
    }

    menu_button_drag = false;
    setTimeout(()=>{
        menu_button_moved = false;
    },0);
});

cap.onclick = event => {
    Menu_Close();
};

btnMenu.onmousedown = event => {
    menu_startPos = [event.clientX, event.clientY];
    menu_button_drag = true;
};

btnMenu.onclick = event => {
    if (menu_button_moved) return;
    Menu_Toogle();
};

searchbox.onclick = event => {
    txtSearch.focus();
};

txtSearch.onkeydown = event => {
    if (event.keyCode == 27) { //esc
        event.stopPropagation();
        if (txtSearch.value.length > 0) {
            txtSearch.value = "";
        } else {
            Menu_Close();
        }
        return;
    }

    if (event.keyCode == 13) { //enter

    }

    if (event.keyCode == 38) { //up
        event.preventDefault();
    }

    if (event.keyCode == 40) { //down
        event.preventDefault();
    }
};

btnSearchClear.onclick = event => {
    event.stopPropagation();

    if (txtSearch.value.length > 0) {
        txtSearch.value = "";
    } else {
        Menu_Close();
    }
};

btnSettings.onclick = event => {
    const winSettings = new Settings();
};

btnLogout.onclick = async event => {
    const f = await fetch("/logout", {
        method: "GET",
        //headers: {"content-type": "application/json"},
        cache: "no-cache",
        credentials: "same-origin"
    });

    console.log(await f.text());
};
