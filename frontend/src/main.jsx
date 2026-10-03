import React, { useEffect, useMemo, useState } from "react";
import { createRoot } from "react-dom/client";
import {
  Check,
  Eye,
  EyeOff,
  ImagePlus,
  LogIn,
  LogOut,
  Mail,
  MapPin,
  Pencil,
  Plus,
  ShoppingBag,
  X
} from "lucide-react";
import "./styles.css";
import { api, authStorage } from "./api";

const USE_MOCK_FILLINGS = false;

const DRAFT_KEY = "caprice_order_draft_v2";
const FILLINGS_CACHE_KEY = "caprice_fillings_cache_v1";

function readLocalJson(key, fallback) {
  try {
    const value = localStorage.getItem(key);
    return value ? JSON.parse(value) : fallback;
  } catch {
    return fallback;
  }
}

const SAVED_DRAFT = readLocalJson(DRAFT_KEY, {});
const SAVED_FILLINGS = readLocalJson(FILLINGS_CACHE_KEY, []);

async function readImageAsDataUrl(file, maxSize = 1280, quality = 0.78) {
  const dataUrl = await new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => resolve(reader.result);
    reader.onerror = reject;
    reader.readAsDataURL(file);
  });

  return new Promise((resolve, reject) => {
    const image = new Image();
    image.onload = () => {
      const scale = Math.min(1, maxSize / Math.max(image.width, image.height));
      const canvas = document.createElement("canvas");
      canvas.width = Math.max(1, Math.round(image.width * scale));
      canvas.height = Math.max(1, Math.round(image.height * scale));
      const ctx = canvas.getContext("2d");
      ctx.drawImage(image, 0, 0, canvas.width, canvas.height);
      resolve({
        name: file.name,
        dataUrl: canvas.toDataURL("image/jpeg", quality)
      });
    };
    image.onerror = reject;
    image.src = dataUrl;
  });
}

const MOCK_FILLINGS = [
  {
    id: "mock-red-velvet",
    name: "Красный бархат",
    price: 2200,
    note: "Нежный шоколадный бисквит с кремом на сливочном сыре.",
    image_url:
      "https://images.unsplash.com/photo-1578985545062-69928b1d9587?auto=format&fit=crop&w=900&q=80",
    is_active: true
  },
  {
    id: "mock-chocolate",
    name: "Шоколадный",
    price: 2400,
    note: "Шоколадный бисквит, насыщенный крем и шоколадная начинка.",
    image_url:
      "https://images.unsplash.com/photo-1606313564200-e75d5e30476c?auto=format&fit=crop&w=900&q=80",
    is_active: true
  },
  {
    id: "mock-strawberry",
    name: "Клубника со сливками",
    price: 2300,
    note: "Ванильный бисквит, сливочный крем и клубничная прослойка.",
    image_url:
      "https://images.unsplash.com/photo-1565958011703-44f9829ba187?auto=format&fit=crop&w=900&q=80",
    is_active: true
  },
  {
    id: "mock-pistachio",
    name: "Фисташковый",
    price: 2600,
    note: "Фисташковый бисквит с нежным кремом и малиновой начинкой.",
    image_url:
      "https://images.unsplash.com/photo-1571115177098-24ec42ed204d?auto=format&fit=crop&w=900&q=80",
    is_active: true
  },
  {
    id: "mock-caramel",
    name: "Солёная карамель",
    price: 2500,
    note: "Шоколадный бисквит, карамельный крем и солёная карамель.",
    image_url:
      "https://images.unsplash.com/photo-1578985545062-69928b1d9587?auto=format&fit=crop&w=900&q=80",
    is_active: true
  },
  {
    id: "mock-mango",
    name: "Манго-маракуйя",
    price: 2700,
    note: "Лёгкий ванильный бисквит с тропической фруктовой начинкой.",
    image_url:
      "https://images.unsplash.com/photo-1551024506-0bccd828d307?auto=format&fit=crop&w=900&q=80",
    is_active: true
  },
  {
    id: "mock-oreo",
    name: "Орео",
    price: 2300,
    note: "Шоколадный бисквит, крем с печеньем Oreo и шоколадная крошка.",
    image_url:
      "https://images.unsplash.com/photo-1602351447937-745cb720612f?auto=format&fit=crop&w=900&q=80",
    is_active: true
  },
  {
    id: "mock-lemon",
    name: "Лимонный",
    price: 2100,
    note: "Ванильный бисквит с лёгким лимонным кремом и цитрусовой прослойкой.",
    image_url:
      "https://images.unsplash.com/photo-1519915028121-7d3463d20b13?auto=format&fit=crop&w=900&q=80",
    is_active: true
  }
];


function formatPrice(value) {
  return new Intl.NumberFormat("ru-RU").format(Number(value) || 0) + " ₽";
}

function decodeJwt(token) {
  try {
    const payload = token.split(".")[1];
    return JSON.parse(atob(payload.replace(/-/g, "+").replace(/_/g, "/")));
  } catch {
    return null;
  }
}

function normalizeFillings(payload) {
  const list = Array.isArray(payload)
    ? payload
    : payload?.fillings ||
      payload?.items ||
      payload?.data ||
      [];

  return list
    .map((item) => ({
      id: item.id ?? item.ID,
      name: item.name ?? item.Name,
      price: Number(
        item.price ??
        item.price_per_kg ??
        item.pricePerKg ??
        item.PricePerKg ??
        item.Price ??
        0
      ),
      note:
        item.description ??
        item.Description ??
        "",
      image_url: item.image_name
        ? `http://localhost:8080/static/fillings/${item.image_name}`
        : "",
      is_active:
        item.is_active ??
        item.isActive ??
        item.IsActive ??
        true
    }))
    .filter((item) => item.id && item.name);
}

function normalizeOrders(payload) {
  const list = Array.isArray(payload)
    ? payload
    : [payload];

  return list.map((item) => {
    const filling = item.fillings?.[0];

    return {
      id: item.id,
      date: item.created_at,
      filling: filling?.name ?? "—",
      weight: Number(item.weight_grams ?? 0) / 1000,
      price: Number(item.total_price ?? 0),
      status: item.status ?? "accepted",
      decoration: item.decoration_wishes ?? "",
      address: item.delivery_address ?? ""
    };
  });
}

function orderTabForStatus(status) {
  if (status === "cancelled") return "cancelled";
  if (status === "ready") return "ready";
  return "processing";
}

function orderStatusLabel(status) {
  return {
    accepted: "Принят",
    processing: "В работе",
    ready: "Готов к выдаче",
    cancelled: "Отменён"
  }[status] || status || "Статус уточняется";
}

function formatOrderDate(value) {
  if (!value) return "Дата заказа";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return String(value);
  return new Intl.DateTimeFormat("ru-RU", {
    day: "numeric",
    month: "long",
    hour: "2-digit",
    minute: "2-digit"
  }).format(date);
}

function App() {
  const [fillings, setFillings] = useState(SAVED_FILLINGS);
  const [selectedFilling, setSelectedFilling] = useState(SAVED_DRAFT.selectedFilling || null);
  const [weight, setWeight] = useState(Number(SAVED_DRAFT.weight) || 1);
  const [serverPrice, setServerPrice] = useState(null);
  const [priceLoading, setPriceLoading] = useState(false);
  const [design, setDesign] = useState(SAVED_DRAFT.design || "");
  const [noDesign, setNoDesign] = useState(Boolean(SAVED_DRAFT.noDesign));
  const [address, setAddress] = useState(SAVED_DRAFT.address || "");
  const [photos, setPhotos] = useState(SAVED_DRAFT.photos || []);
  const [orders, setOrders] = useState([]);
  const [orderTab, setOrderTab] = useState("processing");
  const [ordersOpen, setOrdersOpen] = useState(false);
  const [authOpen, setAuthOpen] = useState(false);
  const [authMode, setAuthMode] = useState("login");
  const [addOpen, setAddOpen] = useState(false);
  const [notice, setNotice] = useState("");
  const [loading, setLoading] = useState(true);
  const [authLoading, setAuthLoading] = useState(false);
  const [paymentOpen, setPaymentOpen] = useState(false);
  const [paymentLoading, setPaymentLoading] = useState(false);
  const [paymentSuccess, setPaymentSuccess] = useState(false);
  const [paymentOrderId, setPaymentOrderId] = useState("");
  const [authUser, setAuthUser] = useState(() => {
    const token = authStorage.getAccess();
    return token ? decodeJwt(token) : null;
  });

  const loggedIn = Boolean(authStorage.getAccess());

  useEffect(() => {
    if (!notice) return undefined;
    const timer = window.setTimeout(() => setNotice(""), 5000);
    return () => window.clearTimeout(timer);
  }, [notice]);

  useEffect(() => {
    try {
      localStorage.setItem(
        DRAFT_KEY,
        JSON.stringify({
          selectedFilling,
          weight,
          design,
          noDesign,
          address,
          photos
        })
      );
    } catch {
      // The draft is a convenience cache; the order itself is still stored on the server.
    }
  }, [selectedFilling, weight, design, noDesign, address, photos]);

  useEffect(() => {
    if (!selectedFilling || !fillings.length) return;
    const fresh = fillings.find((item) => item.id === selectedFilling.id);
    if (fresh && fresh.id !== selectedFilling.id) setSelectedFilling(fresh);
  }, [fillings]);

  const total = useMemo(() => {
    if (!selectedFilling) return 0;
    return Number(selectedFilling.price || 0) * weight + (noDesign || !design.trim() ? 0 : 700);
  }, [selectedFilling, weight, design, noDesign]);

  useEffect(() => {
    loadFillings();
  }, []);

  useEffect(() => {
    if (loggedIn) {
      loadOrders();
    } else {
      setOrders([]);
    }
  }, [loggedIn, authUser?.role]);

async function loadFillings() {
  setLoading(true);
  setNotice("");

  if (USE_MOCK_FILLINGS) {
    setFillings(MOCK_FILLINGS);
    setLoading(false);
    return;
  }

  try {
    const response = await api.getFillings();
    const serverFillings = normalizeFillings(response);
    const cached = readLocalJson(FILLINGS_CACHE_KEY, []);
    const byId = new Map(cached.map((item) => [String(item.id), item]));

    // Keep known hidden fillings in the admin UI even though GET /fillings
    // intentionally returns only active fillings.
    serverFillings.forEach((item) => byId.set(String(item.id), item));
    const merged = Array.from(byId.values());

    setFillings(merged);
    try {
      localStorage.setItem(FILLINGS_CACHE_KEY, JSON.stringify(merged));
    } catch {}
  } catch (error) {
    const cached = readLocalJson(FILLINGS_CACHE_KEY, []);
    setFillings(cached);
    if (!cached.length) {
      setNotice(`Не удалось загрузить начинки: ${error.message}`);
    }
  } finally {
    setLoading(false);
  }
}
  
  async function loadOrders() {
    try {
      const response =
        authUser?.role === "admin"
          ? await api.getAllOrders()
          : await api.getMyOrders();

      const list = Array.isArray(response)
        ? response
        : [];

      console.log("ORDERS FROM LIST:", list);
      const details = await Promise.all(
        list.map(async (order) => {
          try {
            const response = await api.getOrder(order.id ?? order.ID);

            console.log("FULL ORDER:", response);

            return response;
          } catch (error) {
            console.error("GET ORDER ERROR:", error);
            return order;
          }
        })
      );

      setOrders(normalizeOrders(details));
    } catch (error) {
      console.error("Не удалось загрузить заказы:", error);
      setOrders([]);
    }
  }

  async function submitAuth(event) {
    event.preventDefault();

    const form = new FormData(event.currentTarget);
    const email = String(form.get("email") || "").trim();
    const password = String(form.get("password") || "");

    if (!email || !password) {
      setNotice("Заполните email и пароль.");
      return;
    }

    setAuthLoading(true);

    try {
      const tokens = authMode === "register"
        ? await api.register(email, password)
        : await api.login(email, password);

      authStorage.set(tokens);
      setAuthUser(decodeJwt(tokens.access_token));
      setAuthOpen(false);
      setNotice(authMode === "register" ? "Регистрация выполнена." : "Вы вошли в аккаунт.");
      await loadOrders();
    } catch (error) {
      setNotice(error.message);
    } finally {
      setAuthLoading(false);
    }
  }

  async function handleLogout(showNotice = true) {
    await api.logout();
    setAuthUser(null);
    setOrders([]);
    if (showNotice) setNotice("Вы вышли из аккаунта.");
  }

async function placeOrder() {
  if (!selectedFilling) {
    setNotice("Сначала выберите начинку.");
    return;
  }

  if (!loggedIn) {
    setAuthMode("login");
    setAuthOpen(true);
    setNotice(
      "Для оформления заказа сначала войдите или зарегистрируйтесь."
    );
    return;
  }

  if (!address.trim()) {
    setNotice("Укажите адрес доставки.");
    return;
  }

  try {
    const created = await api.createOrder({
      fillings: [
        {
          name: selectedFilling.name,
          weight_grams: Math.round(Number(weight) * 1000)
        }
      ],
      decoration_wishes: noDesign ? "" : design.trim(),
      delivery_address: address.trim()
    });

    const orderId =
      created?.id ??
      created?.ID ??
      created?.order_id ??
      created?.orderId ??
      `TEST-${Date.now().toString().slice(-6)}`;

    setPaymentOrderId(String(orderId));
    setPaymentSuccess(false);
    setPaymentOpen(true);

    await loadOrders();
  } catch (error) {
    setNotice(`Не удалось оформить заказ: ${error.message}`);
  }
}

async function handleMockPayment() {
  setPaymentLoading(true);

  await new Promise((resolve) =>
    setTimeout(resolve, 1200)
  );

  setPaymentLoading(false);
  setPaymentSuccess(true);
}

  async function calculateServerPrice() {
    if (!selectedFilling || !weight) {
      setServerPrice(null);
      return;
    }

    setPriceLoading(true);

    try {
      const response = await api.calculatePrice({
        filling_id: String(selectedFilling.id),
        weight_kg: Number(weight)
      });

      const price =
        response?.amount ??
        response?.Amount ??
        response?.price?.amount ??
        response?.price?.Amount ??
        response?.price ??
        response?.Price;

      if (price !== undefined && price !== null) {
        setServerPrice(Number(price));
      } else {
        setServerPrice(null);
      }
    } catch (error) {
      setServerPrice(null);
      //setNotice(`Не удалось рассчитать цену: ${error.message}`);
    } finally {
      setPriceLoading(false);
    }
  }

  useEffect(() => {
    calculateServerPrice();
  }, [selectedFilling, weight]);

  async function handlePhotosChange(event) {
    const files = Array.from(event.target.files || []).slice(0, 3);
    if (event.target.files?.length > 3) {
      setNotice("Можно прикрепить не более 3 фотографий.");
    }
    const valid = files.filter((file) => file.type.startsWith("image/"));
    if (valid.length !== files.length) {
      setNotice("Прикреплять можно только изображения.");
    }

    try {
      const next = [];
      for (const file of valid) {
        next.push(await readImageAsDataUrl(file));
      }
      setPhotos(next);
    } catch {
      setNotice("Не удалось сохранить одну из фотографий в кэш браузера.");
    }
  }

  function removePhoto(index) {
    setPhotos((current) => current.filter((_, i) => i !== index));
  }

  async function addFilling(event) {
    event.preventDefault();

    const form = new FormData(event.currentTarget);
    const image = form.get("image");

    if (!(image instanceof File) || !image.size) {
      setNotice("Для новой начинки выберите PNG-картинку.");
      return;
    }

    try {
      await api.createFilling(form);
      setAddOpen(false);
      setNotice("Начинка добавлена.");
      await loadFillings();
    } catch (error) {
      setNotice(error.message);
    }
  }

  async function editFilling(filling) {
    const name = window.prompt("Название начинки:", filling.name);
    if (name === null) return;

    const description = window.prompt("Описание:", filling.note || "");
    if (description === null) return;

    const priceText = window.prompt("Цена за кг:", String(filling.price));
    if (priceText === null) return;

    const price = Number(priceText);
    if (!name.trim() || !Number.isFinite(price) || price <= 0) {
      setNotice("Проверьте название и цену.");
      return;
    }

    try {
      const updated = await api.updateFilling(filling.id, {
        name: name.trim(),
        description: description.trim(),
        price
      });

      const normalized = normalizeFillings([updated])[0] || {
        ...filling,
        name: name.trim(),
        note: description.trim(),
        price
      };

      setFillings((current) => {
        const next = current.map((item) =>
          item.id === filling.id ? { ...item, ...normalized, is_active: filling.is_active } : item
        );
        try {
          localStorage.setItem(FILLINGS_CACHE_KEY, JSON.stringify(next));
        } catch {}
        return next;
      });
      setNotice("Начинка обновлена.");
    } catch (error) {
      setNotice(`Не удалось изменить начинку: ${error.message}`);
    }
  }

  async function toggleFilling(filling) {
    try {
      if (filling.is_active) {
        await api.deleteFilling(filling.id);
      } else {
        await api.updateFilling(filling.id, { is_active: true });
      }

      setFillings((current) => {
        const next = current.map((item) =>
          item.id === filling.id ? { ...item, is_active: !filling.is_active } : item
        );
        try {
          localStorage.setItem(FILLINGS_CACHE_KEY, JSON.stringify(next));
        } catch {}
        return next;
      });

      if (selectedFilling?.id === filling.id && filling.is_active) {
        setSelectedFilling(null);
      }

      setNotice(filling.is_active ? "Начинка скрыта." : "Начинка снова доступна.");
    } catch (error) {
      setNotice(`Не удалось изменить доступность: ${error.message}`);
    }
  }

  const filteredOrders = orders.filter((order) => orderTabForStatus(order.status) === orderTab);
  const role = authUser?.role;
  const visibleFillings = role === "admin"
    ? fillings
    : fillings.filter((filling) => filling.is_active !== false);
  const imageFor = (filling) => filling.image_url || "/cake_header.jpg";

  async function changeOrderStatus(orderId, status) {
    try {
      await api.updateOrderStatus(orderId, status);

      setOrders((current) =>
        current.map((order) =>
          order.id === orderId
            ? { ...order, status }
            : order
        )
      );

      setNotice("Статус заказа изменён.");
    } catch (error) {
      setNotice(`Не удалось изменить статус: ${error.message}`);
    }
  }
  
  return (
    <div className="app-shell">
      <main id="top">
        <section className="hero">
          <div className="hero-photo" />
          <div className="hero-label">
            <div className="header-title">CAPRICE</div>
          </div>
        </section>

          <section className="content-section orders-section" id="orders">
            <div className="section-heading">
              <div>
                <p className="eyebrow">Личный кабинет</p>
                <h2>Ваши заказы</h2>
              </div>

              {loggedIn ? (
                <div className="account-actions">
                  <span className="account-email">{authUser?.role === "admin" ? "Администратор" : "Личный кабинет"}</span>
                  <button className="button button-light" onClick={() => handleLogout()}>
                    <LogOut size={17} /> Выйти
                  </button>
                </div>
              ) : (
                <button className="button button-light" onClick={() => {
                  setAuthMode("login");
                  setAuthOpen(true);
                }}>
                  <LogIn size={17} /> Войти
                </button>
              )}
            </div>

            <div className="tabs" role="tablist">
              {[
                ["processing", "Активные"],
                ["ready", "Готовые"],
                ["cancelled", "Отменённые"]
              ].map(([key, label]) => (
                <button
                  key={key}
                  className={orderTab === key ? "tab active" : "tab"}
                  onClick={() => setOrderTab(key)}
                >
                  {label}
                </button>
              ))}
            </div>

            <div className="orders-list">
              {!loggedIn ? (
                <div className="empty-state">
                  Войдите в аккаунт, чтобы увидеть свои заказы.
                </div>
              ) : filteredOrders.length === 0 ? (
                <div className="empty-state">Здесь пока нет заказов.</div>
              ) : (
                filteredOrders.map((order) => (
                  <article className="order-row" key={order.id}>
                    <div className="order-id">
                      <span>Заказ</span>
                      <strong>{order.id}</strong>
                    </div>

                    <div>
                      <span className="order-muted">Дата</span>
                      <strong>{formatOrderDate(order.date)}</strong>
                    </div>

                    <div>
                      <span className="order-muted">Начинка</span>
                      <strong>{order.filling}</strong>
                    </div>

                    <div>
                      <span className="order-muted">Пожелания</span>
                      <strong>{order.decoration || "—"}</strong>
                    </div>

                    <div>
                      <span className="order-muted">Вес</span>
                      <strong>{order.weight} кг</strong>
                    </div>

                    <div>
                      <span className="order-muted">Адрес</span>
                      <strong>{order.address || "—"}</strong>
                    </div>

                    <div className="order-right">
                      <strong className="order-price">
                        {formatPrice(order.price)}
                      </strong>

                      <span className={`order-status-badge status-${order.status}`}>
                        {orderStatusLabel(order.status)}
                      </span>

                      {role === "admin" && (
                        <select
                          className="order-status-select"
                          value={order.status}
                          onChange={(event) =>
                            changeOrderStatus(order.id, event.target.value)
                          }
                        >
                          <option value="processing">В процессе</option>
                          <option value="ready">Готов</option>
                          <option value="cancelled">Отменён</option>
                        </select>
                      )}
                    </div>
                  </article>
                ))
              )}
            </div>
          </section>


        <section className="content-section filling-section" id="filling">
          <div className="section-heading">
              {role === "admin" && (
                <div>
                  <h2>Начинки</h2>
                </div>
              )}
              {role !== "admin" && (
                <div>
                  <h2>Выберите начинку</h2>
                </div>
              )}


            {role === "admin" && (
              <button className="button button-light" onClick={() => setAddOpen(true)}>
                <Plus size={17} /> Добавить начинку
              </button>
            )}
          </div>

          <div className="filling-grid">
            {loading ? (
              <div className="empty-state">Загружаем начинки…</div>
            ) : (
              visibleFillings.map((filling) => (
                <article
                  className={[
                    "filling-card",
                    selectedFilling?.id === filling.id ? "selected" : "",
                    filling.is_active === false ? "inactive" : ""
                  ].filter(Boolean).join(" ")}
                  key={filling.id}
                  role="button"
                  tabIndex={0}
                  onClick={() => setSelectedFilling(filling)}
                  onKeyDown={(event) => {
                    if (event.key === "Enter" || event.key === " ") {
                      event.preventDefault();
                      setSelectedFilling(filling);
                    }
                  }}
                >
                  <div className="filling-image">
                    <img
                      src={imageFor(filling)}
                      alt={filling.name}
                      onError={(event) => {
                        event.currentTarget.src = "/cake_header.jpg";
                      }}
                    />

                  </div>
                  <div className="filling-info">
                    <strong>{filling.name}</strong>
                    <small>{filling.note}</small>
                    <b>{formatPrice(filling.price)} / кг</b>
                    {role === "admin" && (
                      <div className="admin-filling-actions" onClick={(event) => event.stopPropagation()}>
                        <button type="button" className="admin-mini-button" onClick={() => editFilling(filling)}>
                          <Pencil size={14} /> Редактировать
                        </button>
                        <button type="button" className="admin-mini-button" onClick={() => toggleFilling(filling)}>
                          {filling.is_active ? <EyeOff size={14} /> : <Eye size={14} />}
                          {filling.is_active ? "Скрыть" : "Вернуть"}
                        </button>
                      </div>
                    )}
                  </div>
                </article>
              ))
            )}
          </div>
        </section>

        {role !== "admin" && (
          <section className="content-section order-section" id="order">
            <div className="section-heading">
              <div>
                <h2>Оформление заказа</h2>
              </div>
              <div className="selected-summary">
                <span>Начинка</span>
                <strong>{selectedFilling?.name || "Ещё не выбрана"}</strong>
              </div>
            </div>

            {!address.trim() && (
              <div className="order-alert">
                <MapPin size={20} />
                <div>
                  <strong>Укажите адрес доставки</strong>
                  <span>Он понадобится для оформления заказа.</span>
                </div>
              </div>
            )}

            <div className="order-form">
              <div className="form-block">
                <label>Выберите вес</label>
                <div className="weight-picker">
                  {[1, 2, 3, 4, 5, 6, 7].map((value) => (
                    <button
                      type="button"
                      key={value}
                      className={weight === value ? "weight active" : "weight"}
                      onClick={() => setWeight(value)}
                    >
                      {value} кг
                    </button>
                  ))}
                </div>
              </div>

              <div className="form-block">
                <label htmlFor="design">Ваши пожелания по дизайну</label>
                <textarea
                  id="design"
                  value={design}
                  onChange={(e) => setDesign(e.target.value)}
                  placeholder="Например: белый крем, ягоды сверху, надпись «С днём рождения!»"
                  disabled={noDesign}
                />
                <label className="checkbox-row">
                  <input
                    type="checkbox"
                    checked={noDesign}
                    onChange={(e) => setNoDesign(e.target.checked)}
                  />
                  <span>Без дизайна</span>
                </label>
              </div>
              
              <div className="form-block">
                <label htmlFor="references"><ImagePlus size={17} /> Референсы</label>
                <label htmlFor="references" className="reference-upload">
                  <span className="reference-upload__icon">＋</span>
                  <span className="reference-upload__title">
                    Добавить референсы
                  </span>
                </label>

                <input
                  id="references"
                  className="reference-input"
                  type="file"
                  accept="image/*"
                  multiple
                />
                <small className="field-hint">Можно прикрепить от 0 до 3 фото.</small>
                {photos.length > 0 && (
                  <div className="reference-grid">
                    {photos.map((photo, index) => (
                      <div className="reference-item" key={`${photo.name}-${index}`}>
                        <img src={photo.dataUrl} alt={photo.name} />
                        <button type="button" onClick={() => removePhoto(index)} aria-label="Удалить фото">
                          <X size={14} />
                        </button>
                      </div>
                    ))}
                  </div>
                )}
              </div>

              <div className="form-block">
                <label htmlFor="address"><MapPin size={17} /> Адрес доставки</label>
                <input
                  id="address"
                  value={address}
                  onChange={(e) => setAddress(e.target.value)}
                  placeholder="Город, улица, дом, квартира"
                />
              </div>

              <div className="checkout-card">
                <div>
                  <span>Итоговая цена</span>

                  <strong className="checkout-price">
                    {priceLoading
                      ? "Рассчитываем…"
                      : serverPrice !== null
                        ? formatPrice(serverPrice)
                        : selectedFilling
                          ? formatPrice(total)
                          : "—"}
                  </strong>
                </div>

                <div className="checkout-actions">
                  <button
                    className="button button-accent checkout-button"
                    onClick={placeOrder}
                    disabled={!selectedFilling || priceLoading}
                  >
                    Оформить заказ
                    <ShoppingBag size={18} />
                  </button>
                </div>
              </div>
            </div>
          </section>
        )}
      </main>

      <footer className="site-footer">
        <div className="footer-inner">

          <div className="footer-brand">
            <h3>Каприз — кондитерская, Москва</h3>
            <p>
              Свежие торты к вашему празднику
            </p>
          </div>

          <div className="footer-column">
            <h4>Контакты</h4>

            <p>
              <a href="tel:+79001209267">
                +7 (900) 120-92-67
              </a>
            </p>

            <p>
              <a href="mailto:info@kapriz.ru">
                info@kapriz.ru
              </a>
            </p>
          </div>

          <div className="footer-column">
            <h4>Адрес</h4>

            <p>
              Москва, улица Тортиков, дом 1
            </p>

            <h4>Часы работы</h4>

            <p>
              ежедневно, 10:00 – 22:00
            </p>
          </div>

          <div className="footer-column footer-wide">
            <div className="footer-cancel-box">
              <h4><Mail size={16} /> Хотите отменить заказ?</h4>
              <p>Напишите нам на <a href="mailto:info@kapriz.ru">info@kapriz.ru</a>, указав номер заказа.</p>
            </div>
            <h4>Информация</h4>

            <p className="footer-links">
              <a href="#">
                Политика конфиденциальности
              </a>

              <span> · </span>

              <a href="#">
                Согласие на обработку ПД
              </a>
            </p>

            <p className="allergens">
              <strong>Аллергены:</strong>{" "}
              продукция может содержать следы
              глютена, молока, яиц, орехов, сои,
              кунжута и какао. Состав указан
              в карточках товаров.
            </p>

            <p className="allergy-warning">
              При аллергии сообщите нам до
              оформления заказа.
            </p>
          </div>

        </div>

        <div className="footer-bottom">
          <span>
            ООО «Каприз», ИНН 7727676767, Москва
          </span>

          <span>
            © 2025 Каприз. Все права защищены.
          </span>
        </div>
      </footer>

      {notice && (
        <div className="toast toast-visible">
          <Check size={18} />
          <span>{notice}</span>
          <button onClick={() => setNotice("")}><X size={17} /></button>
        </div>
      )}

      {authOpen && (
        <Modal
          title={authMode === "login" ? "Войти в личный кабинет" : "Создать аккаунт"}
          onClose={() => setAuthOpen(false)}
        >
          <div className="auth-switch">
            <button
              className={authMode === "login" ? "auth-switch-active" : ""}
              onClick={() => setAuthMode("login")}
            >
              Вход
            </button>
            <button
              className={authMode === "register" ? "auth-switch-active" : ""}
              onClick={() => setAuthMode("register")}
            >
              Регистрация
            </button>
          </div>

          <form onSubmit={submitAuth}>
            <label className="modal-label" htmlFor="email">Email</label>
            <input className="modal-input" id="email" name="email" type="email" autoComplete="email" placeholder="you@example.com" required />

            <label className="modal-label" htmlFor="password">Пароль</label>
            <input className="modal-input" id="password" name="password" type="password" minLength="3" autoComplete={authMode === "login" ? "current-password" : "new-password"} placeholder="Минимум 3 символа" required />

            <button className="button button-accent modal-submit" disabled={authLoading}>
              {authLoading ? "Подождите…" : authMode === "login" ? "Войти" : "Зарегистрироваться"}
            </button>
          </form>
        </Modal>
      )}

      {paymentOpen && (
        <Modal
          title={paymentSuccess ? "Спасибо за заказ!" : "Оплата заказа"}
          onClose={() => {
            if (!paymentLoading) {
              setPaymentOpen(false);
            }
          }}
        >
          {!paymentSuccess ? (
            <>
              <p>
                Заказ <strong>#{paymentOrderId}</strong> создан.
              </p>

              <p>
                Это тестовая оплата. Деньги реально не списываются.
              </p>

              <button
                className="button button-accent modal-submit"
                onClick={handleMockPayment}
                disabled={paymentLoading}
              >
                {paymentLoading ? "Обрабатываем оплату…" : "Оплатить"}
              </button>
            </>
          ) : (
            <>
              <p>
                Спасибо за заказ!
              </p>

              <p>
                Номер заказа: <strong>#{paymentOrderId}</strong>
              </p>

              <button
                className="button button-accent modal-submit"
                onClick={() => setPaymentOpen(false)}
              >
                Закрыть
              </button>
            </>
          )}
        </Modal>
      )}

      {addOpen && (
        <Modal title="Добавить начинку" onClose={() => setAddOpen(false)}>
          <form onSubmit={addFilling}>
            <label className="modal-label" htmlFor="filling-name">Название</label>
            <input className="modal-input" id="filling-name" name="name" placeholder="Фисташка" required />

            <label className="modal-label" htmlFor="filling-description">Описание</label>
            <input className="modal-input" id="filling-description" name="description" placeholder="Фисташковый крем" />

            <label className="modal-label" htmlFor="filling-price">Цена за кг</label>
            <input className="modal-input" id="filling-price" name="price" type="number" min="0" placeholder="1400" required />

            <label className="modal-label" htmlFor="filling-image">Изображение (PNG)</label>
            <input className="modal-file" id="filling-image" name="image" type="file" accept="image/png" required />
            <button className="button button-accent modal-submit">Добавить на сервер</button>
          </form>
        </Modal>
      )}
    </div>
  );
}

function Modal({ title, children, onClose }) {
  return (
    <div className="modal-backdrop" onMouseDown={onClose}>
      <div className="modal" onMouseDown={(e) => e.stopPropagation()}>
        <button className="modal-close" onClick={onClose}><X size={20} /></button>
        <h3>{title}</h3>
        {children}
      </div>
    </div>
  );
}

createRoot(document.getElementById("root")).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
);
