package main

import "fmt"

const BaseDeliveryRate = 10.0
const MaxBatteryLevel = 100.0
const EmergencyThreshold = 20.0
const BufferCapacity = 5

type RobotType string

const (
	Drone       RobotType = "Drone"
	Wheeled     RobotType = "Wheeled"
	HeavyLifter RobotType = "HeavyLifter"
)

var allowedRobotTypes = []RobotType{Drone, Wheeled, HeavyLifter}

type Product struct {
	ID     int
	Name   string
	Weight float64
}

type Robot struct {
	ID          int
	Model       string
	Type        RobotType
	Battery     float64
	IsAvailable bool
}

type Mover interface {
	Move(distance float64) error
}

func (robot *Robot) Move(distance float64) error {
	rate := 0.0
	switch robot.Type {
	case Drone:
		rate = 1.5
	case Wheeled:
		rate = 0.8
	case HeavyLifter:
		rate = 2.5
	default:
		return fmt.Errorf("неизвестный тип робота")
	}
	needed := distance * rate
	if robot.Battery < needed {
		return fmt.Errorf("недостаточно заряда")
	}
	robot.Battery -= needed
	return nil
}

func Recharge(robot *Robot) {
	robot.Battery = MaxBatteryLevel
}

func canCarry(robotType RobotType, weight float64) bool {
	maxWeight := 0.0
	switch robotType {
	case Drone:
		maxWeight = 2
	case Wheeled:
		maxWeight = 10
	case HeavyLifter:
		maxWeight = 50
	}
	return weight <= maxWeight
}

func trimSpaces(value string) string {
	start := 0
	end := len(value)
	for start < end && (value[start] == ' ' || value[start] == '\t' || value[start] == '\n' || value[start] == '\r') {
		start++
	}
	for end > start && (value[end-1] == ' ' || value[end-1] == '\t' || value[end-1] == '\n' || value[end-1] == '\r') {
		end--
	}
	return value[start:end]
}

func DeliveryTask(robot *Robot, product Product, distance float64, reports chan<- string) {
	steps := int(distance * BaseDeliveryRate)
	for i := 0; i < steps; i++ {
		_ = i
	}
	if err := robot.Move(distance); err != nil {
		reports <- fmt.Sprintf("ОШИБКА: заказ %d, робот %d: %v", product.ID, robot.ID, err)
		return
	}
	reports <- fmt.Sprintf("УСПЕШНО: заказ %d доставлен роботом %d", product.ID, robot.ID)
}

func main() {
	warehouse := map[string]int{
		"Книга":   5,
		"Коробка": 4,
		"Деталь":  6,
	}
	productWeights := map[string]float64{
		"Книга":   1.0,
		"Коробка": 5.0,
		"Деталь":  15.0,
	}
	robots := []Robot{
		{ID: 1, Model: "Sky", Type: Drone, Battery: 90, IsAvailable: true},
		{ID: 2, Model: "Roller", Type: Wheeled, Battery: 80, IsAvailable: true},
		{ID: 3, Model: "Lift", Type: HeavyLifter, Battery: 100, IsAvailable: true},
	}
	var orderBuffer [BufferCapacity]Product
	orderCount := 0
	nextProductID := 1
	_ = allowedRobotTypes

	for {
		fmt.Println("\n1. Добавить товар в заказ")
		fmt.Println("2. Показать склад и роботов")
		fmt.Println("3. Отправить заказы на доставку")
		fmt.Println("4. Зарядить робота")
		fmt.Println("5. Выход")
		fmt.Print("Выберите пункт: ")
		var choice int
		if _, err := fmt.Scan(&choice); err != nil {
			fmt.Println("Ошибка ввода")
			return
		}

		switch choice {
		case 1:
			if orderCount >= BufferCapacity {
				fmt.Println("Буфер заказов заполнен")
				continue
			}
			fmt.Print("Введите название товара без пробелов: ")
			var name string
			if _, err := fmt.Scan(&name); err != nil {
				fmt.Println("Ошибка ввода")
				continue
			}
			name = trimSpaces(name)
			if warehouse[name] <= 0 {
				fmt.Println("Товара нет на складе")
				continue
			}
			warehouse[name]--
			orderBuffer[orderCount] = Product{ID: nextProductID, Name: name, Weight: productWeights[name]}
			nextProductID++
			orderCount++
			fmt.Println("Товар добавлен в заказ")

		case 2:
			fmt.Println("Склад:")
			for name, amount := range warehouse {
				fmt.Printf("%s: %d\n", name, amount)
			}
			fmt.Println("Роботы:")
			for i := range robots {
				availability := "занят"
				if robots[i].IsAvailable {
					availability = "свободен"
				}
				fmt.Printf("ID %d, %s, тип %s, заряд %.1f, %s\n", robots[i].ID, robots[i].Model, robots[i].Type, robots[i].Battery, availability)
			}

		case 3:
			if orderCount == 0 {
				fmt.Println("Нет заказов")
				continue
			}
			fmt.Print("Введите расстояние доставки: ")
			var distance float64
			if _, err := fmt.Scan(&distance); err != nil || distance < 0 {
				fmt.Println("Некорректное расстояние")
				continue
			}
			var nextBuffer [BufferCapacity]Product
			nextCount := 0
			reports := make(chan string, orderCount)
			tasks := 0
			for orderIndex := 0; orderIndex < orderCount; orderIndex++ {
				product := orderBuffer[orderIndex]
				robotIndex := -1
				for i := range robots {
					if robots[i].IsAvailable && robots[i].Battery > EmergencyThreshold && canCarry(robots[i].Type, product.Weight) {
						robotIndex = i
						break
					}
				}
				if robotIndex == -1 {
					nextBuffer[nextCount] = product
					nextCount++
					continue
				}
				robots[robotIndex].IsAvailable = false
				tasks++
				go DeliveryTask(&robots[robotIndex], product, distance, reports)
			}
			for i := 0; i < tasks; i++ {
				fmt.Println(<-reports)
			}
			close(reports)
			for i := range robots {
				if !robots[i].IsAvailable {
					robots[i].IsAvailable = true
				}
			}
			orderBuffer = nextBuffer
			orderCount = nextCount
			fmt.Printf("Осталось необработанных заказов: %d\n", orderCount)

		case 4:
			fmt.Print("Введите ID робота: ")
			var id int
			if _, err := fmt.Scan(&id); err != nil {
				fmt.Println("Ошибка ввода")
				continue
			}
			found := false
			for i := range robots {
				if robots[i].ID == id {
					Recharge(&robots[i])
					fmt.Println("Робот заряжен")
					found = true
					break
				}
			}
			if !found {
				fmt.Println("Робот не найден")
			}

		case 5:
			fmt.Println("Программа завершена")
			return

		default:
			fmt.Println("Нет такого пункта меню")
		}
	}
}
