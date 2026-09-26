import React, { useEffect, useMemo, useState } from "react";
import { createRoot } from "react-dom/client";
import {
  Check,
  LogIn,
  LogOut,
  MapPin,
  Plus,
  ShoppingBag,
  User,
  X
} from "lucide-react";
import "./styles.css";
import { api, authStorage } from "./api";

const USE_MOCK_FILLINGS = false;

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
    .filter(
      (item) =>
        item.id &&
        item.name &&
        item.is_active !== false
    );
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
  if (status === "delivered") return "done";
  return "active";
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
  const [fillings, setFillings] = useState([]);
  const [selectedFilling, setSelectedFilling] = useState(null);
  const [weight, setWeight] = useState(1);
  const [serverPrice, setServerPrice] = useState(null);
  const [priceLoading, setPriceLoading] = useState(false);
  const [design, setDesign] = useState("");
  const [noDesign, setNoDesign] = useState(false);
  const [address, setAddress] = useState("");
  const [orders, setOrders] = useState([]);
  const [orderTab, setOrderTab] = useState("active");
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

    setFillings(serverFillings);

    //if (!serverFillings.length) {
      //setNotice("На сервере пока нет начинок.");
    //}
  } catch (error) {
    setFillings([]);
    //setNotice(
     // `Не удалось загрузить начинки с сервера: ${error.message}`
    //);
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

  async function addFilling(event) {
    event.preventDefault();

    const form = new FormData(event.currentTarget);
    const payload = {
      name: String(form.get("name") || "").trim(),
      description: String(form.get("description") || "").trim(),
      price: Number(form.get("price") || 0),
      image_name: String(form.get("image_name") || "").trim()
    };

    try {
      await api.createFilling(payload);
      setAddOpen(false);
      setNotice("Начинка отправлена на сервер.");
      await loadFillings();
    } catch (error) {
      setNotice(error.message);
    }
  }

  const filteredOrders = orders.filter((order) => orderTabForStatus(order.status) === orderTab);
  const role = authUser?.role;
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
                ["active", "Активные"],
                ["done", "Выполненные"],
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
            <div>
              <h2>Выберите начинку</h2>
            </div>

            {/* {role === "admin" && (
              <button className="button button-light" onClick={() => setAddOpen(true)}>
                <Plus size={17} /> Добавить начинку
              </button>
            )} */}
          </div>

          <div className="filling-grid">
            {loading ? (
              <div className="empty-state">Загружаем начинки…</div>
            ) : (
              fillings.map((filling) => (
                <button
                  className={selectedFilling?.id === filling.id ? "filling-card selected" : "filling-card"}
                  key={filling.id}
                  onClick={() => setSelectedFilling(filling)}
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
                  </div>
                </button>
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
        <div className="toast">
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

            <label className="modal-label" htmlFor="filling-image">Имя файла картинки</label>
            <input className="modal-input"id="filling-image" name="image_name" placeholder="filling1.jpg"/>
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
